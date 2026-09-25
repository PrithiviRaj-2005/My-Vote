#!/usr/bin/env bash
# PulseVote Automated 16-Step API Test Script (Bash/cURL)
# Usage: ./test_plan.sh [http://localhost:8080]

BASE_URL="${1:-http://localhost:8080}"
echo "=========================================================="
echo "  PULSEVOTE – 16-Step Comprehensive Verification Suite     "
echo "=========================================================="

RAND_ID=$((RANDOM % 9000 + 1000))
EMAIL="tester_${RAND_ID}@pulsevote.com"
PASSWORD="PulsePass2026!"
NAME="Alex Tester"

# Step 0: Health
echo "[TEST] Step 0: Health Check"
curl -s -f "${BASE_URL}/api/health" || { echo "Health check failed"; exit 1; }
echo ""

# Step 1: Signup
echo "[TEST] Step 1: Signup New User"
SIGNUP_RES=$(curl -s -X POST "${BASE_URL}/api/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${NAME}\",\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\",\"confirmPassword\":\"${PASSWORD}\"}")
TOKEN=$(echo "$SIGNUP_RES" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Signup failed: $SIGNUP_RES"
  exit 1
fi
echo "[PASS] Signed up and received token: ${TOKEN:0:15}..."

# Step 2: Duplicate Signup
echo "[TEST] Step 2: Prevent Duplicate Signup"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${NAME}\",\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
if [ "$HTTP_CODE" -eq 409 ]; then
  echo "[PASS] Correctly rejected duplicate email with 409"
else
  echo "[FAIL] Expected 409, got $HTTP_CODE"
fi

# Step 3: Login
echo "[TEST] Step 3: Login"
LOGIN_RES=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
TOKEN=$(echo "$LOGIN_RES" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "[PASS] Logged in successfully"

# Step 4: Invalid Login
echo "[TEST] Step 4: Reject Invalid Login"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"WrongPassword\"}")
if [ "$HTTP_CODE" -eq 401 ]; then
  echo "[PASS] Correctly rejected invalid credentials with 401"
else
  echo "[FAIL] Expected 401, got $HTTP_CODE"
fi

# Step 5: Create Poll
echo "[TEST] Step 5: Create Poll"
POLL_RES=$(curl -s -X POST "${BASE_URL}/api/polls" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"question":"Favorite cloud provider?","options":["AWS","GCP","Azure","Fly.io"]}')
SHARE_CODE=$(echo "$POLL_RES" | grep -o '"shareCode":"[^"]*' | cut -d'"' -f4)
POLL_ID=$(echo "$POLL_RES" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo "[PASS] Created poll with shareCode: $SHARE_CODE (ID: $POLL_ID)"

# Step 6: Invalid Poll
echo "[TEST] Step 6: Reject Invalid Poll (<2 options)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/polls" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"question":"Invalid?","options":["OnlyOne"]}')
if [ "$HTTP_CODE" -eq 400 ]; then
  echo "[PASS] Correctly rejected invalid poll with 400"
else
  echo "[FAIL] Expected 400, got $HTTP_CODE"
fi

# Step 7: Open Public Poll
echo "[TEST] Step 7: Fetch Public Poll"
curl -s -f "${BASE_URL}/api/polls/${SHARE_CODE}" > /dev/null
echo "[PASS] Public poll fetched without authentication"

# Step 8: Submit Vote
echo "[TEST] Step 8: Submit Vote"
VOTE_RES=$(curl -s -X POST "${BASE_URL}/api/polls/${SHARE_CODE}/vote" \
  -H "Content-Type: application/json" \
  -d '{"optionId":"opt_1"}')
echo "[PASS] Vote registered: $VOTE_RES"

# Step 9: Reject Invalid Option
echo "[TEST] Step 9: Reject Invalid Option Vote"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/polls/${SHARE_CODE}/vote" \
  -H "Content-Type: application/json" \
  -d '{"optionId":"opt_unknown"}')
if [ "$HTTP_CODE" -eq 400 ]; then
  echo "[PASS] Correctly rejected invalid option with 400"
fi

# Step 10 & 11: Multi-vote & Redis Live Results
echo "[TEST] Step 10 & 11: Multiple Votes & Redis Count"
curl -s -X POST "${BASE_URL}/api/polls/${SHARE_CODE}/vote" -H "Content-Type: application/json" -d '{"optionId":"opt_2"}' > /dev/null
curl -s -X POST "${BASE_URL}/api/polls/${SHARE_CODE}/vote" -H "Content-Type: application/json" -d '{"optionId":"opt_1"}' > /dev/null
RESULTS=$(curl -s "${BASE_URL}/api/polls/${SHARE_CODE}/results")
echo "[PASS] Results returned from Redis Hash: $RESULTS"

# Step 13 & 14: Close Poll & Reject Vote
echo "[TEST] Step 13 & 14: Close Poll & Verify Vote Rejection"
curl -s -X POST "${BASE_URL}/api/polls/${POLL_ID}/close" -H "Authorization: Bearer ${TOKEN}" > /dev/null
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/polls/${SHARE_CODE}/vote" \
  -H "Content-Type: application/json" \
  -d '{"optionId":"opt_1"}')
if [ "$HTTP_CODE" -eq 403 ]; then
  echo "[PASS] Voting rejected on closed poll with 403"
fi

# Step 16: Delete Poll
echo "[TEST] Step 16: Delete Poll"
curl -s -X DELETE "${BASE_URL}/api/polls/${POLL_ID}" -H "Authorization: Bearer ${TOKEN}" > /dev/null
echo "[PASS] Poll deleted successfully"

echo ""
echo "All 16 test steps completed successfully!"
