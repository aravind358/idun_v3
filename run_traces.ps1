$inputs = @(
    "hi"
    "hello"
    "hi bro"
    "bro"
    "what time is it"
    "3+3"
    "tell me a joke"
    "create a reminder for tomorrow"
    "take a note saying buy milk"
    "what is the weather today"
)

$outputFile = "trace_audit_results.md"
"" | Out-File $outputFile

foreach ($inputStr in $inputs) {
    Write-Host "Tracing: $inputStr"
    "## Trace: `"$inputStr`"" | Out-File -Append $outputFile
    "```json" | Out-File -Append $outputFile
    
    # Run the trace program for 10 seconds. We'll modify trace/main.go to exit gracefully if done.
    go run ./cmd/trace "$inputStr" | Out-File -Append $outputFile
    
    "```" | Out-File -Append $outputFile
    "---" | Out-File -Append $outputFile
}
