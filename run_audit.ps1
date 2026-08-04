$inputs = @(
    # POSITIVE INPUTS
    "hi",
    "hello",
    "hey",
    "good morning",
    "bye",
    "thanks",
    "thank you",
    "yes",
    "okay",
    "no",
    "cancel",
    "who are you",
    "how are you",
    "Delete report.pdf",
    "remove old_logs.txt",
    "open notes.txt",
    "rename document.docx",
    "move file doc.txt to archive",
    "create directory new_folder",
    "list files in downloads",
    "3+3",
    "what is 15 plus 27",
    "calculate 100/5",
    "divide 10 by 2",
    "multiply 4 and 5",
    "(15+20)*3",
    "12.5 * 8",
    "remind me to call John tomorrow",
    "set a reminder to study for exam at 8 PM",
    "weather in Tokyo",
    "what is the weather today",
    "temperature tomorrow",
    "will it rain next Friday",
    "forecast for today",
    "take a note saying buy milk",
    "save this that password is admin",
    "remember to buy eggs",
    "delete the note",
    "open that note",
    "what time is it",
    "tell me the time",
    "date today",
    "battery level",
    "cpu usage",
    "memory usage",
    "how much disk space",
    "shutdown",
    "restart",
    "lock screen",

    # NEGATIVE INPUTS
    "Delete happiness",
    "What is the meaning of life?",
    "Calculate the weight of a black hole",
    "Weather in Narnia",
    "Remind my cat to meow",
    "Open my mind",

    # AMBIGUOUS INPUTS
    "Delete it",
    "Open that",
    "Remind me",
    "Take a note",

    # INVALID INPUTS
    "Delete",
    "Calculate",
    "Weather in",
    "Rename",
    "Remind to",

    # EDGE CASES
    "Can you tell me the time please",
    "Delete report.pdf to C:\temp",
    "what day is it today"
)

Remove-Item -Path "robustness_audit_raw2.txt" -ErrorAction Ignore
go build -o trace.exe ./cmd/trace

foreach ($input in $inputs) {
    echo "--- INPUT: $input ---" >> robustness_audit_raw2.txt
    ./trace.exe "$input" >> robustness_audit_raw2.txt 2>&1
}
