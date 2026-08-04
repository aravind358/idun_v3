# Slot Extraction Audit Report

| Input | Intent | Extracted Slots | Status |
|---|---|---|---|
| divide 100 by 5 | N/A | None | FAIL |
| multiply 8 and 12 | N/A | None | FAIL |
| add 15 and 20 | N/A | None | FAIL |
| subtract 5 from 10 | N/A | None | FAIL |
| calculate 15 plus 20 | N/A | None | FAIL |
| what is 15 * 20 | N/A | None | FAIL |
| solve 10 / 2 | N/A | None | FAIL |
| (15+20)*3 | N/A | None | FAIL |
| divide by zero | N/A | None | FAIL |
| multiply something | N/A | None | FAIL |
| calculate the meaning of life | N/A | None | FAIL |
| Remind John tomorrow at 5 PM to call Sarah | N/A | None | FAIL |
| set a reminder tomorrow at 5 PM to call John | N/A | None | FAIL |
| remind me to buy milk next Monday at 9 AM | N/A | None | FAIL |
| remind me in 2 hours to check the oven | N/A | None | FAIL |
| remind us to leave in 15 minutes | N/A | None | FAIL |
| remind me to pay bills tomorrow | N/A | None | FAIL |
| remind me to feed the dog at 8 AM | N/A | None | FAIL |
| remind Sarah to submit the report | N/A | None | FAIL |
| set reminder to take out trash | N/A | None | FAIL |
| remind me to call John about the meeting tomorrow at 5 PM | N/A | None | FAIL |
| remind me to tell Jane to email Mark in 3 hours | N/A | None | FAIL |
| weather in Tokyo tomorrow | query_weather | city=`Tokyo tomorrow` | PASS |
| what is the forecast next Friday in New York | N/A | None | FAIL |
| temperature in London for the next 3 days | query_weather | city=`London for the next 3 days` | PASS |
| what is the weather for the next 5 days in Paris | N/A | None | FAIL |
| weather in Berlin | query_weather | city=`Berlin` | PASS |
| temperature tomorrow | query_weather | city=`tomorrow` | PASS |
| forecast for the next 2 days | query_weather | city=`for the next 2 days` | PASS |
| will it rain in Seattle tomorrow | N/A | None | FAIL |
| will it rain in Seattle | N/A | None | FAIL |
| will it rain next Monday | N/A | None | FAIL |
| weather | query_weather | None | PASS |
| will it rain | N/A | None | FAIL |
| move report.pdf to C:/archive | N/A | None | FAIL |
| copy document.docx to /home/user/backup | N/A | None | FAIL |
| rename notes.txt to old_notes.txt | N/A | None | FAIL |
| open file C:/projects/notes.txt | N/A | None | FAIL |
| delete old_data.csv | N/A | None | FAIL |
| create directory C:/new_folder | N/A | None | FAIL |
| mkdir test_folder | N/A | None | FAIL |
| list files in C:/projects | N/A | None | FAIL |
| list files | N/A | None | FAIL |
| move my super long filename with spaces.txt to /var/log | N/A | None | FAIL |
| open report | N/A | None | FAIL |
| take a note called Ideas saying build a robot | N/A | None | FAIL |
| create a note titled Shopping List saying buy milk and eggs | N/A | None | FAIL |
| save this named Meeting Notes saying project is delayed | N/A | None | FAIL |
| remember that I left the keys on the table | N/A | None | FAIL |
| take a note saying call mom | N/A | None | FAIL |
| delete that note called Ideas | N/A | None | FAIL |
| remove the note | N/A | None | FAIL |
| open note titled Shopping List | N/A | None | FAIL |
| read the note | N/A | None | FAIL |
| take a note | N/A | None | FAIL |
| take a note called Empty | N/A | None | FAIL |
| what time is it | N/A | None | FAIL |
| date today | N/A | None | FAIL |
| battery level | N/A | None | FAIL |
| how much ram | N/A | None | FAIL |
| how much disk space | N/A | None | FAIL |
| shutdown computer tomorrow | N/A | None | FAIL |
| turn off system | N/A | None | FAIL |
| restart system next Friday | N/A | None | FAIL |
| reboot | N/A | None | FAIL |
| lock screen | N/A | None | FAIL |
| lock computer | N/A | None | FAIL |
| shutdown the toaster | N/A | None | FAIL |
| how much water | N/A | None | FAIL |
