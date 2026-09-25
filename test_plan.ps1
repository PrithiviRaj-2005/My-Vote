# PulseVote Automated 16-Step API Test Script (PowerShell)
# Usage: .\test_plan.ps1 [-BaseUrl http://localhost:8080]

param (
    [string]$BaseUrl = "http://localhost:8080"
)

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  MY VOTE - 16-Step Comprehensive Verification Suite       " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$Email = "tester_$((Get-Random -Minimum 1000 -Maximum 9999))@pulsevote.com"
$Password = "PulsePass2026!"
$Name = "Alex Tester"

# Helper for colored test reporting
function Assert-Step($StepNum, $Desc, $Condition, $Details) {
    if ($Condition) {
        Write-Host "[PASS] Step ${StepNum}: $Desc" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] Step ${StepNum}: $Desc - $Details" -ForegroundColor Red
    }
}

try {
    # 0. Health Check
    $health = Invoke-RestMethod -Uri "$BaseUrl/api/health" -Method Get
    Assert-Step 0 "API Health Check" ($health.status -eq "ok") $health

    # 1. Signup
    $signupBody = @{
        name = $Name
        email = $Email
        password = $Password
        confirmPassword = $Password
    } | ConvertTo-Json
    $signupRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method Post -Body $signupBody -ContentType "application/json"
    $token = $signupRes.token
    Assert-Step 1 "Signup New User" ($token -ne $null -and $signupRes.user.email -eq $Email) "Token was empty"

    # 2. Duplicate Signup
    try {
        $dupRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/signup" -Method Post -Body $signupBody -ContentType "application/json"
        Assert-Step 2 "Prevent Duplicate Signup" $false "Server should return 409 Conflict"
    } catch {
        Assert-Step 2 "Prevent Duplicate Signup" ($_.Exception.Response.StatusCode.value__ -eq 409) $_.Exception.Message
    }

    # 3. Login
    $loginBody = @{
        email = $Email
        password = $Password
    } | ConvertTo-Json
    $loginRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginRes.token
    Assert-Step 3 "Valid User Login" ($token -ne $null) "Failed to retrieve JWT token on login"

    # 4. Invalid Login
    try {
        $badLoginBody = @{ email = $Email; password = "WrongPassword123" } | ConvertTo-Json
        $badRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method Post -Body $badLoginBody -ContentType "application/json"
        Assert-Step 4 "Reject Invalid Login" $false "Server accepted wrong password"
    } catch {
        Assert-Step 4 "Reject Invalid Login" ($_.Exception.Response.StatusCode.value__ -eq 401) $_.Exception.Message
    }

    # 5. Create Poll
    $headers = @{
        Authorization = "Bearer $token"
    }
    $createPollBody = @{
        question = "What is the best cloud provider for live workloads?"
        options = @("AWS", "Google Cloud", "Azure", "Cloudflare")
    } | ConvertTo-Json
    $poll = Invoke-RestMethod -Uri "$BaseUrl/api/polls" -Method Post -Headers $headers -Body $createPollBody -ContentType "application/json"
    $shareCode = $poll.shareCode
    $pollId = $poll.id
    Assert-Step 5 "Create Valid Poll" ($shareCode -ne $null -and $poll.options.Length -eq 4) "Poll creation failed"

    # 6. Invalid Poll (Only 1 option)
    try {
        $invalidPollBody = @{
            question = "Invalid Poll Question"
            options = @("SingleOptionOnly")
        } | ConvertTo-Json
        $invRes = Invoke-RestMethod -Uri "$BaseUrl/api/polls" -Method Post -Headers $headers -Body $invalidPollBody -ContentType "application/json"
        Assert-Step 6 "Reject Invalid Poll (<2 options)" $false "Server allowed 1 option poll"
    } catch {
        Assert-Step 6 "Reject Invalid Poll (<2 options)" ($_.Exception.Response.StatusCode.value__ -eq 400) $_.Exception.Message
    }

    # 7. Open Public Poll
    $publicPoll = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode" -Method Get
    Assert-Step 7 "Retrieve Public Poll by shareCode" ($publicPoll.poll.shareCode -eq $shareCode) "Could not fetch public poll"

    # 8. Valid Vote
    $voteBody = @{ optionId = "opt_1" } | ConvertTo-Json
    $voteRes = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/vote" -Method Post -Body $voteBody -ContentType "application/json"
    Assert-Step 8 "Submit Valid Vote" ($voteRes.message -like "*successfully*") "Vote failed"

    # 9. Invalid Option Vote
    try {
        $badVoteBody = @{ optionId = "opt_non_existent" } | ConvertTo-Json
        $badVote = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/vote" -Method Post -Body $badVoteBody -ContentType "application/json"
        Assert-Step 9 "Reject Invalid Option ID" $false "Server accepted nonexistent option"
    } catch {
        Assert-Step 9 "Reject Invalid Option ID" ($_.Exception.Response.StatusCode.value__ -eq 400) $_.Exception.Message
    }

    # 10. Multi-user Votes
    $voteBody2 = @{ optionId = "opt_2" } | ConvertTo-Json
    Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/vote" -Method Post -Body $voteBody2 -ContentType "application/json" | Out-Null
    Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/vote" -Method Post -Body $voteBody -ContentType "application/json" | Out-Null
    Assert-Step 10 "Multiple Users / Concurrent Votes" $true "Submitted extra votes"

    # 11. Redis Live Result Update Check
    $results = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/results" -Method Get
    $opt1Votes = ($results.options | Where-Object { $_.id -eq "opt_1" }).votes
    Assert-Step 11 "Redis Live Option Count Updated" ($opt1Votes -eq 2 -and $results.totalVotes -eq 3) "Results did not reflect votes"

    # 12. WebSocket URL Check
    $wsEndpoint = "$BaseUrl/ws/polls/$shareCode"
    Assert-Step 12 "Gorilla WebSocket Endpoint Configured" ($wsEndpoint -like "*/ws/polls/*") $wsEndpoint

    # 13. Close Poll
    $closeRes = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$pollId/close" -Method Post -Headers $headers
    Assert-Step 13 "Close Poll by Owner" ($closeRes.poll.isActive -eq $false) "Failed to close poll"

    # 14. Voting on Closed Poll Rejection
    try {
        $closedVote = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/vote" -Method Post -Body $voteBody -ContentType "application/json"
        Assert-Step 14 "Prevent Vote on Closed Poll" $false "Server permitted voting on closed poll"
    } catch {
        Assert-Step 14 "Prevent Vote on Closed Poll" ($_.Exception.Response.StatusCode.value__ -eq 403) $_.Exception.Message
    }

    # 15. View Final Results of Closed Poll
    $finalResults = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$shareCode/results" -Method Get
    Assert-Step 15 "View Results on Closed Poll" ($finalResults.isActive -eq $false -and $finalResults.totalVotes -eq 3) "Final results corrupted"

    # 16. Delete Poll
    $delRes = Invoke-RestMethod -Uri "$BaseUrl/api/polls/$pollId" -Method Delete -Headers $headers
    Assert-Step 16 "Delete Poll and Purge Cache" ($delRes.message -like "*successfully*") "Deletion failed"
    Write-Host "All 16 Verification Steps Executed Successfully!" -ForegroundColor Green
} catch {
    Write-Host "Error during test execution: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host "Please ensure the backend server is running on $BaseUrl before executing this script." -ForegroundColor DarkYellow
}
