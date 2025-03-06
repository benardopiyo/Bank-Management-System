# Bank Management System

## Introduction

- This project presents a comprehensive Bank Management System, meticulously crafted using the Golang programming language. It serves as a demonstration of robust software development practices, emphasizing scalability and maintainability. 
- By leveraging Golang's inherent strengths, such as concurrency and efficient memory management, the system aims to provide a practical and performant solution for managing banking operations.


## Features:

**User Authentication:**
- Secure user registration and login functionality.
- Implementation of unique user constraints to prevent duplicate registrations.

**Account Management:**
- Create and manage various account types (Savings, Fixed-Term, Current).
- Update account holder information (e.g., contact details).
- Close existing accounts.
- Transfer account ownership.
    
**Account Information and Transactions:**
- Retrieve detailed account information, including balances and transaction history.
- Support for deposit and withdrawal transactions on eligible accounts.
- Display interest information based on account type.
            - Savings: 7%
            Fixed-Term 1 (fixed01): 4%
            Fixed-Term 2 (fixed02): 5%
            Fixed-Term 3 (fixed03): 8%
            Current: No interest.
**Implement transaction restrictions for fixed-term accounts.
    Data Persistence:
        Store and retrieve user and account data using a suitable data storage mechanism (e.g., files, database).
        Persist transaction logs.

Implemented Enhancements (Golang Specific):

    Concurrency:
        Leverage Golang's concurrency features (goroutines, channels) to handle concurrent user requests and transactions efficiently.
    Error Handling:
        Implement robust error handling throughout the system, providing informative error messages.
    Data Validation:
        Enforce data validation rules to ensure data integrity.
    Modular Design:
        Structure the application using a modular design to improve maintainability and testability.

### Installation (Golang):

- Ensure Golang is installed on your machine.

- Clone the repository:

```
git clone https://github.com/benardopiyo/Bank-Management-System
```

- Navigate to the project directory:

```
cd Bank-Management-System
```

- Run the application:

```
go run ./cmd/web
```