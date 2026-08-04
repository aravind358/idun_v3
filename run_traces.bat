@echo off
set OUT=trace_audit_results.md
echo. > %OUT%

for %%I in ("hi" "hello" "hi bro" "bro" "what time is it" "3+3" "tell me a joke" "create a reminder for tomorrow" "take a note saying buy milk" "what is the weather today") do (
    echo Tracing: %%~I
    echo ## Trace: "%%~I" >> %OUT%
    echo ```json >> %OUT%
    go run ./cmd/trace "%%~I" >> %OUT% 2>&1
    echo ``` >> %OUT%
    echo --- >> %OUT%
)
