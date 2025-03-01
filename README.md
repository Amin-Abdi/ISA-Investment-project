# ISA-Investment-project


## Introduction
The purpose of this report is to outline the development of a backend API service designed to support Individual Savings Account (ISA) investments for retail customers using the Gin framework in Golang. 
Currently, X offers ISA investments exclusively to employees whose employers have an existing arrangement with the company. This project expands X’s offering to direct retail customers, allowing them to invest independently without employer affiliation.

The primary goal of this solution is to provide a secure, scalable, and efficient API that enables customers to:
- Open a X ISA account.
- Select a single fund from a list of available investment options.
- Invest a specified amount into the chosen fund.
- Retrieve details of their ISA, investment history, and available funds.

## Solution Overview
The backend API provides a set of endpoints that allow retail customers to interact with Individual Savings Accounts (ISAs) and Funds. The API is built using Golang, with the Gin framework handling HTTP requests and PostgreSQL as the database.

The API is encapsulated in the `Server` struct:

```go
type Server struct {
	Store StoreInterface
}
```
StoreInterface abstracts database interactions, making the API more modular and testable.

The server defines several routes to manage ISAs, funds, and investments.

### ISA Management
| Method | Endpoint       | Description            |
|--------|----------------|------------------------|
| `POST` | `/isa`         | Create a new ISA       |
| `GET`  | `/isa/:id`     | Retrieve ISA details   |

I have designed the system to allow for future flexibility by supporting multiple fund selections for an ISA, even though customers are currently restricted to selecting just one fund. By using an array to store fund IDs, I ensure that the system can easily be adapted in the future to handle multiple fund options. For now, I have implemented a check to ensure that no fund is already associated with an ISA before adding a new one.

### Fund Management
| Method | Endpoint                       | Description                  |
|--------|--------------------------------|------------------------------|
| `POST` | `/fund`                        | Create a new fund            |
| `GET`  | `/funds`                       | List all funds               |
| `PUT`  | `/funds/:id`                   | Update fund details          |
| `PUT`  | `/isa/:isa_id/fund/:fund_id`   | Associate a fund with an ISA |

I have limited fund updates to only the name and description to avoid potential issues with critical details like risk level, performance, or total amount being altered. Allowing full updates could create legal, compliance, and financial risks. The engineering team should work with legal and finance to define which fund details can be changed and under what conditions.

Additionally, fund management endpoints are admin-only and should not be accessible to customers. Proper access controls must be in place to restrict these functionalities to authorised personnel.

### Investments
| Method | Endpoint                      | Description                              |
|--------|-------------------------------|------------------------------------------|
| `POST` | `/isa/:id/invest`             | Invest into a selected fund              |
| `GET`  | `/investments/:isa_id`        | Retrieve all investments for an ISA      |

I did not implement a delete investment endpoint to ensure compliance with UK financial regulations, which require maintaining transaction records for auditing purposes. Deleting investment data could compromise the integrity of the audit trail and violate regulations such as those from the FCA and HMRC. Keeping all records ensures transparency, protects clients, and supports compliance with anti-money laundering (AML) and know-your-customer (KYC) standards.

The API uses logrus for structured logging, ensuring traceability and providing a detailed log of every action. 

### Mocks
I used **mocking for the API layer** and a **real database for the Postgres layer** to balance isolated unit testing with integration testing. Mocking the API layer allowed for fast, focused tests of business logic, request validation, and error handling without relying on a real database, ensuring quick feedback and precise control over test data. It also enabled easy simulation of error conditions and external services. 

In contrast, using a real database for the Postgres layer was essential for validating actual database interactions, such as query execution, data consistency, and handling complex relationships and constraints. This approach ensured that both the API and database layers were tested in realistic environments, guaranteeing their functionality while maintaining test efficiency.


## Assumptions

I assumed that user authentication would not be handled within this service, as this is primarily an investment service rather than an account management system. It is generally not advisable to mix different domains into a single service, as it can lead to unnecessary complexity and violations of the single responsibility principle. 

Authentication and user management would ideally be handled by a separate service, ensuring better scalability, maintainability, and security. By focusing on investments, this service can remain specialised and streamlined, with the assumption that other systems will handle user identity verification and authorization.

I assumed that the list of available funds would be pre-configured and managed by administrators. The API does not support real-time updates to available funds based on market changes, liquidity issues, or other external factors. Future updates may be required to handle fund availability dynamically.

## Enhancements

### Use of Kafka
If I had more time, I would integrate Kafka to enable an event-driven architecture for better scalability, reliability, and real-time processing. Currently, the API handles transactions synchronously, meaning all operations (such as investment creation, and logging) happen within the request-response cycle. This can introduce delays and make the system less resilient to failures.

By introducing Kafka, I could decouple these processes by publishing events to a message queue, allowing different services to consume them asynchronously. For example:
- **Investment Event Processing**: When a user makes an investment, the API could publish an InvestmentCreated event to a Kafka topic. Consumers such as a notification service (for email confirmations), a compliance service (for regulatory reporting), and a fraud detection service could process these events independently.
- **Real-Time Fund Updates**: Kafka could be used to stream real-time updates on fund performance or price changes, allowing customers to see the latest information without requiring expensive database queries.

### Role Based Access Control
Another key improvement I could add is Role-Based Access Control (RBAC) to enforce proper authorsation for different endpoints. Currently, the API assumes all users have equal access, but in a real-world scenario, different roles should have different permissions.
- **Customers** should be able to create and manage their own ISAs, make investments, and view their portfolio.
- **Admins** should have exclusive permissions to create, update, and manage funds, ensuring that only authorised personnel can modify critical investment options.

This could be implemented using middleware that checks the user's role before processing a request. Tools like Casbin (a Golang authorisation library) or JWT claims could be used to enforce role-based restrictions efficiently.

### Metrics with Grafana and Prometheus
I would also integrate Prometheus for metrics collection and Grafana for visualisation to track crucial operational and financial data. This would provide real-time insights into system performance and investment trends.

Some key metrics I would like to track include:

1. Total money invested per fund over time – to monitor fund popularity and performance.
2. Number of investment transactions per hour/day – to detect activity spikes or anomalies.
3. API request latency and error rates – to ensure system health and reliability.
4. Cash balance trends across all ISAs – to analyze customer behavior and liquidity.

By implementing these metrics, I could gain better visibility into the system, proactively detect issues, and provide valuable analytics for both technical monitoring and business decision-making.

## API Documentation
You can view the OpenAPI spec here:  
[openapi.json](./docs/api.json)
