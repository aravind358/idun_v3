# Slot Extraction Audit Report

| Input | Intent | Extracted Slots | Status |
|---|---|---|---|
| divide 100 by 5 | calculate | operator=`divide`, operand1=`100`, operand2=`5` | PASS |
| multiply 8 and 12 | calculate | operator=`multiply`, operand1=`8`, operand2=`12` | PASS |
| add 15 and 20 | calculate | operator=`add`, operand1=`15`, operand2=`20` | PASS |
| subtract 5 from 10 | calculate | operator=`subtract`, operand2=`5`, operand1=`10` | PASS |
| calculate 15 plus 20 | calculate | operand1=`15`, operator=`plus`, operand2=`20` | PASS |
| what is 15 * 20 | calculate | operand1=`15`, operator=`*`, operand2=`20` | PASS |
| solve 10 / 2 | calculate | operand1=`10`, operator=`/`, operand2=`2` | PASS |
| (15+20)*3 | calculate | expression=`(15+20)*3` | PASS |
| divide by zero | N/A | None | FAIL |
| multiply something | N/A | None | FAIL |
| calculate the meaning of life | calculate | expression=`the meaning of life` | PASS |
| Remind John tomorrow at 5 PM to call Sarah | create_reminder | person=`John`, date=`tomorrow`, time=`5 PM`, task=`call Sarah` | PASS |
| set a reminder tomorrow at 5 PM to call John | create_reminder | date=`tomorrow`, time=`5 PM`, task=`call John` | PASS |
| remind me to buy milk next Monday at 9 AM | create_reminder | target=`me`, task=`buy milk`, date=`next Monday`, time=`9 AM` | PASS |
| remind me in 2 hours to check the oven | create_reminder | person=`me`, duration=`2 hours`, task=`check the oven` | PASS |
| remind us to leave in 15 minutes | create_reminder | target=`us`, task=`leave`, duration=`15 minutes` | PASS |
| remind me to pay bills tomorrow | create_reminder | target=`me`, task=`pay bills`, date=`tomorrow` | PASS |
| remind me to feed the dog at 8 AM | create_reminder | target=`me`, task=`feed the dog`, time=`8 AM` | PASS |
| remind Sarah to submit the report | create_reminder | person=`Sarah`, task=`submit the report` | PASS |
| set reminder to take out trash | create_reminder | task=`take out trash` | PASS |
| remind me to call John about the meeting tomorrow at 5 PM | create_reminder | target=`me`, task=`call John about the meeting`, date=`tomorrow`, time=`5 PM` | PASS |
| remind me to tell Jane to email Mark in 3 hours | create_reminder | target=`me`, task=`tell Jane to email Mark`, duration=`3 hours` | PASS |
| weather in Tokyo tomorrow | query_weather | location=`Tokyo`, date=`tomorrow` | PASS |
| what is the forecast next Friday in New York | calculate | expression=`the forecast next Friday in New York` | PASS |
| temperature in London for the next 3 days | query_weather | location=`London`, duration=`next 3 days` | PASS |
| what is the weather for the next 5 days in Paris | calculate | expression=`the weather for the next 5 days in Paris` | PASS |
| weather in Berlin | query_weather | location=`Berlin` | PASS |
| temperature tomorrow | query_weather | date=`tomorrow` | PASS |
| forecast for the next 2 days | query_weather | duration=`next 2 days` | PASS |
| will it rain in Seattle tomorrow | query_weather | location=`Seattle`, date=`tomorrow` | PASS |
| will it rain in Seattle | query_weather | location=`Seattle` | PASS |
| will it rain next Monday | query_weather | date=`next Monday` | PASS |
| weather | query_weather | None | PASS |
| will it rain | query_weather | None | PASS |
| move report.pdf to C:/archive | N/A | None | FAIL |
| copy document.docx to /home/user/backup | N/A | None | FAIL |
| rename notes.txt to old_notes.txt | file_operation | operation=`rename`, source=`notes.txt`, filename=`notes.txt`, extension=`txt`, destination=`old_notes.txt` | PASS |
| open file C:/projects/notes.txt | N/A | None | FAIL |
| delete old_data.csv | file_operation | operation=`delete`, source=`old_data.csv`, filename=`old_data.csv`, extension=`csv` | PASS |
| create directory C:/new_folder | N/A | None | FAIL |
| mkdir test_folder | create_directory | directory=`test_folder` | PASS |
| list files in C:/projects | N/A | None | FAIL |
| list files | list_files | None | PASS |
| move my super long filename with spaces.txt to /var/log | N/A | None | FAIL |
| open report | file_operation | operation=`open`, source=`report`, filename=`report` | PASS |
| take a note called Ideas saying build a robot | take_note | title=`Ideas`, content=`build a robot` | PASS |
| create a note titled Shopping List saying buy milk and eggs | take_note | title=`Shopping List`, content=`buy milk and eggs` | PASS |
| save this named Meeting Notes saying project is delayed | take_note | title=`Meeting Notes`, content=`project is delayed` | PASS |
| remember that I left the keys on the table | take_note | content=`I left the keys on the table` | PASS |
| take a note saying call mom | take_note | content=`call mom` | PASS |
| delete that note called Ideas | file_operation | operation=`delete`, source=`that note called Ideas`, filename=`that note called Ideas` | PASS |
| remove the note | file_operation | operation=`remove`, source=`the note`, filename=`the note` | PASS |
| open note titled Shopping List | file_operation | operation=`open`, source=`note titled Shopping List`, filename=`note titled Shopping List` | PASS |
| read the note | file_operation | operation=`read`, source=`the note`, filename=`the note` | PASS |
| take a note | N/A | None | FAIL |
| take a note called Empty | take_note | title=`Empty` | PASS |
| what time is it | query_time | None | PASS |
| date today | query_date | None | PASS |
| battery level | query_battery | target=`battery` | PASS |
| how much ram | query_memory | target=`ram` | PASS |
| how much disk space | query_disk | target=`disk` | PASS |
| shutdown computer tomorrow | system_shutdown | target=`computer`, date=`tomorrow` | PASS |
| turn off system | system_shutdown | target=`system` | PASS |
| restart system next Friday | system_restart | target=`system`, date=`next Friday` | PASS |
| reboot | N/A | None | FAIL |
| lock screen | system_lock | target=`screen` | PASS |
| lock computer | system_lock | target=`computer` | PASS |
| shutdown the toaster | N/A | None | FAIL |
| how much water | N/A | None | FAIL |
