# logger
A log management utility for go appications.

Logger provides a log manager that writes log messages to a file. The LogManager can write to more than one log type by 
adding a "FileLogger" to the manager. The user can specify the log location, and the base filename. For rotating logs, the logger appends
a volume number (e.g. "001", "002", "003".

The currently avaialble log file types are:
1. Basic log file. Appends to the end of the log file indefintely. 
2. Limed log file. When created, a high water mark for the file size is specified. Once the high water mark is reached, it rotates the file.
3. Timed log file. When created, the rotation frequency is specified (e.g. 1 hour). At the designated time, the log file rotates. 
4. Daily log file. This file rotates each day at midnight. 

The library also provides a log formatter. Several predefined formats are available. A developer can also create and assign their own
formatter and assign it to a LogFile instance. 
