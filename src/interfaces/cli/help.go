/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import "fmt"

// showHelp displays help information
// Показывает справочную информацию
func showHelp() {
	fmt.Println("Atom Engine - BPMN Process Engine")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd start           - Start daemon in background")
	fmt.Println("  atomd run             - Start daemon in foreground")
	fmt.Println("  atomd stop            - Stop running daemon")
	fmt.Println("  atomd status          - Show daemon status")
	fmt.Println("  atomd events          - Show system events from database")
	fmt.Println("  atomd storage <cmd>   - Storage management commands")
	fmt.Println("  atomd timer <cmd>     - Timer management commands")
	fmt.Println("  atomd process <cmd>   - Process management commands")
	fmt.Println("  atomd token <cmd>     - Token management commands")
	fmt.Println("  atomd job <cmd>       - Job management commands")
	fmt.Println("  atomd message <cmd>   - Message management commands")
	fmt.Println("  atomd expression <cmd> - Expression evaluation commands")
	fmt.Println("  atomd bpmn <cmd>      - BPMN management commands")
	fmt.Println("  atomd help            - Show this help")
	fmt.Println("")
	fmt.Println("Storage commands:")
	fmt.Println("  atomd storage status  - Show storage status")
	fmt.Println("  atomd storage info    - Show storage information")
	fmt.Println("")
	fmt.Println("Timer commands:")
	fmt.Println("  atomd timer add       - Add new timer (ISO 8601 duration)")
	fmt.Println("  atomd timer remove    - Remove timer")
	fmt.Println("  atomd timer status    - Get timer status")
	fmt.Println("  atomd timer list      - List all timers")
	fmt.Println("  atomd timer stats     - Show timewheel statistics")
	fmt.Println("")
	fmt.Println("Process commands:")
	fmt.Println("  atomd process start   - Start process instance")
	fmt.Println("  atomd process status  - Get process instance status")
	fmt.Println("  atomd process cancel  - Cancel process instance")
	fmt.Println("  atomd process list    - List process instances")
	fmt.Println("")
	fmt.Println("Token commands:")
	fmt.Println("  atomd token list      - List all tokens")
	fmt.Println("  atomd token show      - Show token details")
	fmt.Println("  atomd token trace     - Trace token execution path")
	fmt.Println("")
	fmt.Println("Job commands:")
	fmt.Println("  atomd job create      - Create new job")
	fmt.Println("  atomd job list        - List jobs")
	fmt.Println("  atomd job show        - Show job details")
	fmt.Println("  atomd job activate    - Activate jobs for worker")
	fmt.Println("  atomd job complete    - Complete job")
	fmt.Println("  atomd job fail        - Fail job")
	fmt.Println("  atomd job cancel      - Cancel job")
	fmt.Println("  atomd job stats       - Show job statistics")
	fmt.Println("")
	fmt.Println("Message commands:")
	fmt.Println("  atomd message list       - List buffered messages")
	fmt.Println("  atomd message subscriptions - List message subscriptions")
	fmt.Println("  atomd message stats      - Show message statistics")
	fmt.Println("  atomd message publish    - Publish message")
	fmt.Println("  atomd message cleanup    - Cleanup expired messages")
}

// showStorageHelp displays storage help information
// Показывает справочную информацию по storage
func showStorageHelp() {
	fmt.Println("Storage management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd storage status  - Show storage status")
	fmt.Println("  atomd storage info    - Show storage information and statistics")
	fmt.Println("  atomd storage help    - Show this help")
}

// showTimerHelp displays timer help information
// Показывает справочную информацию по timer
func showTimerHelp() {
	fmt.Println("Timer management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd timer add <id> <duration_or_cycle>                  - Add new timer")
	fmt.Println("  atomd timer remove <id>                                   - Remove timer")
	fmt.Println("  atomd timer status <id>                                   - Get timer status")
	fmt.Println("  atomd timer list [status] [limit]                         - List all timers")
	fmt.Println("  atomd timer stats                                         - Show timewheel statistics")
	fmt.Println("  atomd timer help                                          - Show this help")
	fmt.Println("")
	fmt.Println("Duration formats (ISO 8601):")
	fmt.Println("  PT30S                                                     - 30 seconds")
	fmt.Println("  PT5M                                                      - 5 minutes")
	fmt.Println("  PT1H                                                      - 1 hour")
	fmt.Println("  P1D                                                       - 1 day")
	fmt.Println("  P1DT2H30M                                                 - 1 day 2 hours 30 minutes")
	fmt.Println("")
	fmt.Println("Repeating cycles (ISO 8601):")
	fmt.Println("  R5/PT30S                                                  - Repeat 5 times every 30 seconds")
	fmt.Println("  R/PT1M                                                    - Repeat infinitely every minute")
	fmt.Println("  R3/PT10S                                                  - Repeat 3 times every 10 seconds")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd timer add timer1 PT5S                              - Add timer for 5 seconds")
	fmt.Println("  atomd timer add timer2 R3/PT10S                          - Repeat 3 times every 10 seconds")
	fmt.Println("  atomd timer add timer3 R/PT30M                           - Repeat infinitely every 30 minutes")
	fmt.Println("  atomd timer remove timer1                                - Remove timer1")
	fmt.Println("  atomd timer status timer1                                - Check timer1 status")
	fmt.Println("  atomd timer list                                          - List all timers")
	fmt.Println("  atomd timer list SCHEDULED                               - List only scheduled timers")
	fmt.Println("  atomd timer list \"\" 10                                    - List first 10 timers")
}

// showProcessHelp displays process help information
// Показывает справочную информацию по process
func showProcessHelp() {
	fmt.Println("Process management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd process start <process_key> [-v version] [-d variables]  - Start process instance")
	fmt.Println("  atomd process status <instance_id>                             - Get process instance status")
	fmt.Println("  atomd process cancel <instance_id> [reason]                    - Cancel process instance")
	fmt.Println("  atomd process list [status] [limit]                           - List process instances")
	fmt.Println("  atomd process help                                             - Show this help")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -v, --version <version>                                        - Specific version to start")
	fmt.Println("  -d, --data <json>                                              - Process variables as JSON")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd process start Process_Big_Process_ID                     - Start latest version")
	fmt.Println("  atomd process start Process_Big_Process_ID -v 3                - Start version 3")
	fmt.Println("  atomd process start Process_Big_Process_ID -d '{\"data\": \"value\"}'  - Start with variables")
	fmt.Println("  atomd process status srv1-aB3dEf9hK2mN5pQ8uV                  - Get instance status")
	fmt.Println("  atomd process cancel srv1-aB3dEf9hK2mN5pQ8uV \"user requested\"  - Cancel with reason")
	fmt.Println("  atomd process list                                             - List all instances")
	fmt.Println("  atomd process list ACTIVE                                      - List only active instances")
	fmt.Println("  atomd process list \"\" 10                                       - List first 10 instances")
}

// showTokenHelp displays token help information
// Показывает справочную информацию по token
func showTokenHelp() {
	fmt.Println("Token management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd token list [instance_id] [state]                        - List tokens")
	fmt.Println("  atomd token show <token_id>                                   - Show token details")
	fmt.Println("  atomd token trace <instance_id>                               - Trace token execution path")
	fmt.Println("  atomd token help                                              - Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd token list                                              - List all tokens")
	fmt.Println("  atomd token list atom-8Gb7OM3URKioy78glT                      - List tokens for process instance")
	fmt.Println("  atomd token list \"\" ACTIVE                                    - List only active tokens")
	fmt.Println("  atomd token show atom-tokenid12345                           - Show specific token")
	fmt.Println("  atomd token trace atom-8Gb7OM3URKioy78glT                     - Trace execution path")
}

// showJobHelp displays job help information
// Показывает справочную информацию по job
func showJobHelp() {
	fmt.Println("Job management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd job list [type] [worker] [limit]                        - List jobs")
	fmt.Println("  atomd job show <job_key>                                      - Show job details")
	fmt.Println("  atomd job activate <type> <worker> [-j max_jobs] [-t timeout] - Activate jobs for worker")
	fmt.Println("  atomd job complete <job_key> [variables]                      - Complete job")
	fmt.Println("  atomd job fail <job_key> <retries> [error] [backoff]          - Fail job")
	fmt.Println("  atomd job cancel <job_key>                                    - Cancel job")
	fmt.Println("  atomd job help                                                - Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd job list                                                - List all jobs")
	fmt.Println("  atomd job list service-task                                   - List service-task jobs")
	fmt.Println("  atomd job list \"\" worker1                                     - List jobs for worker1")
	fmt.Println("  atomd job show atom-jobkey12345                               - Show specific job")
	fmt.Println("  atomd job activate service-task worker1                       - Activate 1 job (default)")
	fmt.Println("  atomd job activate service-task worker1 -j 5                  - Activate up to 5 jobs")
	fmt.Println("  atomd job activate service-task worker1 -t 5000               - Activate job with 5s timeout")
	fmt.Println("  atomd job activate service-task worker1 -j 3 -t 10000         - Activate 3 jobs with 10s timeout")
	fmt.Println("  atomd job complete atom-jobkey12345 '{\"result\": \"success\"}'  - Complete with variables")
	fmt.Println("  atomd job fail atom-jobkey12345 2 \"Connection failed\"         - Fail with 2 retries left")
	fmt.Println("  atomd job cancel atom-jobkey12345                             - Cancel job")
}

// showMessageHelp displays message help information
// Показывает справочную информацию по message
func showMessageHelp() {
	fmt.Println("Message management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd message publish <name> [correlation_key] [variables] [ttl]  - Publish message")
	fmt.Println("  atomd message list [tenant_id] [limit]                           - List correlation results")
	fmt.Println("  atomd message subscriptions [tenant_id] [limit]                  - List subscriptions")
	fmt.Println("  atomd message buffered [tenant_id] [limit]                       - List buffered messages")
	fmt.Println("  atomd message cleanup [tenant_id]                                - Cleanup expired messages")
	fmt.Println("  atomd message stats [tenant_id]                                  - Show message statistics")
	fmt.Println("  atomd message help                                               - Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd message publish order_created                              - Publish simple message")
	fmt.Println("  atomd message publish order_created order123                     - Publish with correlation key")
	fmt.Println("  atomd message publish order_created order123 '{\"amount\": 100}'  - Publish with variables")
	fmt.Println("  atomd message list                                               - List all messages")
	fmt.Println("  atomd message subscriptions                                      - List all subscriptions")
	fmt.Println("  atomd message buffered                                           - List buffered messages")
	fmt.Println("  atomd message cleanup                                            - Cleanup expired messages")
	fmt.Println("  atomd message stats                                              - Show statistics")
}

// showExpressionHelp displays expression help information
// Показывает справочную информацию по expression
func showExpressionHelp() {
	fmt.Println("Expression management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd expression eval <expression> [context]                     - Evaluate expression")
	fmt.Println("  atomd expression validate <expression> [schema]                  - Validate expression")
	fmt.Println("  atomd expression parse <expression>                              - Parse expression to AST")
	fmt.Println("  atomd expression functions [category]                            - List supported functions")
	fmt.Println("  atomd expression test <expression> <test_cases>                  - Test expression")
	fmt.Println("  atomd expression help                                            - Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd expression eval \"x + y\" '{\"x\": 5, \"y\": 3}'              - Evaluate arithmetic")
	fmt.Println("  atomd expression eval \"user.age > 18\" '{\"user\": {\"age\": 25}}'  - Evaluate condition")
	fmt.Println("  atomd expression validate \"x + y\"                               - Validate syntax")
	fmt.Println("  atomd expression parse \"count(items)\"                           - Parse to AST")
	fmt.Println("  atomd expression functions string                                - List string functions")
	fmt.Println("  atomd expression functions                                       - List all functions")
}

// showBPMNHelp displays BPMN help information
// Показывает справочную информацию по bpmn
func showBPMNHelp() {
	fmt.Println("BPMN management commands:")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  atomd bpmn parse <file.bpmn> [process_id] [--force|-f]   - Parse BPMN file")
	fmt.Println("  atomd bpmn list [limit]                                   - List all BPMN processes")
	fmt.Println("  atomd bpmn show <process_key>                             - Show BPMN process details (use PROCESS KEY from list)")
	fmt.Println("  atomd bpmn delete <process_id>                            - Delete BPMN process")
	fmt.Println("  atomd bpmn stats                                          - Show BPMN statistics")
	fmt.Println("  atomd bpmn json <process_key>                             - Show process JSON data (use PROCESS KEY from list)")
	fmt.Println("  atomd bpmn help                                           - Show this help")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atomd bpmn parse process.bpmn                             - Parse process.bpmn")
	fmt.Println("  atomd bpmn parse process.bpmn my-process-1                - Parse with specified ID")
	fmt.Println("  atomd bpmn parse process.bpmn --force                     - Force import")
	fmt.Println("  atomd bpmn parse process.bpmn my-process-1 -f             - Force with ID")
	fmt.Println("  atomd bpmn list                                           - List all processes")
	fmt.Println("  atomd bpmn list 10                                        - List first 10 processes")
	fmt.Println("  atomd bpmn show atom-7-1k2-PVn4Y9j-CF5M                   - Show details (PROCESS KEY)")
	fmt.Println("  atomd bpmn delete my-process-1                            - Delete process")
	fmt.Println("  atomd bpmn stats                                          - Show parser statistics")
	fmt.Println("  atomd bpmn json atom-7-1k2-PVn4Y9j-CF5M                   - Show JSON data (PROCESS KEY)")
}
