$inputs = @(
    "3+3"
    "what time is it"
    "what is the date"
    "create a folder called test1"
    "battery status"
)

$outputFile = "presentation_adoption_report.md"
"# Presentation Pipeline Certification Report" | Out-File $outputFile

foreach ($inputStr in $inputs) {
    Write-Host "Tracing: $inputStr"
    "## Trace: $inputStr" | Out-File -Append $outputFile
    '```json' | Out-File -Append $outputFile
    
    go run ./cmd/trace "$inputStr" | Out-File -Append $outputFile
    
    '```' | Out-File -Append $outputFile
    "---" | Out-File -Append $outputFile
}
