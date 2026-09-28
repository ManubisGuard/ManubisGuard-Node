.

## Project Structure
```
├───backend             # Backend handler and interfaces
│   └───xray            # Xray methods and jobs
│       └───api         # Xray API handler
├───common              # Proto files and common object structures
├───config              # Reads .env configuration
├───controller          # Service controllers for managing API interactions  
│   ├───rest            # REST API protocol methods
│   └───rpc             # gRPC protocol methods
├───logger              # primary logger for backend logs
└───tools               # Standalone utilities with no project dependencies
```
