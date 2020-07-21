# simple-wallet - Demo golang app

Sequence of actions when starting the application

1. Create database using `schema.up.sql` from `sql/` folder
2. Export environment variable `DATABASE_URL`
```bash
$ export DATABASE_URL=postgres://{user}:{password}@{host}:{port}/{db_name}
```
3.  Build and run golang app
```bash
$ cd ./cmd/wallet/
$ go build 
$ ./wallet
```

## Example Usage
### Create wallet

REQUEST `POST /wallet/create`

```bash
POST /wallet/create
{
    "firstname" : "Askhat",
    "lastname" : "Kaparov"
}
```
RESPONSE  `{wallet-id}`  - 32 symbol (a-z0-9)
```bash
HTTP/1.1 201 CREATED
{
"id" : "rb86mmhupg4z6apdli3yk2y579a2zhyq"
}
```


example:
```bash
$ curl -i -X POST -H 'Content-Type: application/json' -d '{"firstname": "Askhat", "lastname": "Kaparov"}' http://0.0.0.0:8080/wallet/create
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Date: Tue, 21 Jul 2020 07:18:57 GMT
Content-Length: 41

{"id":"rb86mmhupg4z6apdli3yk2y579a2zhyq"}
```
### Topup wallet

REQUEST `POST /wallet/topup`

```bash
POST /wallet/topup
{
    "recipient-id": "xyk4lzbuh8n9gswnb7pvnhgbn5imkygl",
    "amount": 1000
}
```
RESPONSE  `{transaction-id}`  - integer
```bash
HTTP/1.1 201 CREATED
{
"transaction-id" : 1342
}
```


example:
```bash
$ curl -i -X POST -H 'Content-Type: application/json' -d '{"recipient-id": "xyk4lzbuh8n9gswnb7pvnhgbn5imkygl", "amount": 1000}' http://0.0.0.0:8080/wallet/topup
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Date: Tue, 21 Jul 2020 07:26:39 GMT
Content-Length: 21

{"transaction-id":75}
```

### Send from wallet to wallet

REQUEST `POST /wallet/send`

```bash
POST /wallet/send
{
    "sender-id": "nfr06q05wg74ne5yr95kwsi1bm4fl0eq",
    "recipient-id": "xyk4lzbuh8n9gswnb7pvnhgbn5imkygl",
    "amount": 1000
}
```
RESPONSE  `{transaction-id}`  - integer
```bash
HTTP/1.1 201 CREATED
{
"transaction-id" : 64
}
```


example:

```bash
$ curl -i -X POST -H 'Content-Type: application/json' -d '{"sender-id":"nfr06q05wg74ne5yr95kwsi1bm4fl0eq", "recipient-id": "ccj7b1oz91uc8vm7z5qjyctrjjpdb81g", "amount": 5}' http://0.0.0.0:8080/wallet/send
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Date: Tue, 21 Jul 2020 07:29:05 GMT
Content-Length: 21

{"transaction_id":76}
```
### Get wallet activity history given wallet id, date and direction


REQUEST `GET /wallet/{id}/history?date={dd-mm-yyyy}&direction=[deposit|withdraw]`

```bash
GET /wallet/nfr06q05wg74ne5yr95kwsi1bm4fl0eq/history?date=20-07-2020&direction=deposit
```
RESPONSE  csv like string
```bash
HTTP/1.1 200 OK
transaction-id,sender-id,recipient-id,amount,time
transaction-id,sender-id,recipient-id,amount,time
transaction-id,sender-id,recipient-id,amount,time
transaction-id,sender-id,recipient-id,amount,time
...
transaction-id,sender-id,recipient-id,amount,time

```


example:


```bash
$ curl -i -X GET "http://0.0.0.0:8080/wallet/nfr06q05wg74ne5yr95kwsi1bm4fl0eq/history?date=20-07-2020&direction=deposit"
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Tue, 21 Jul 2020 07:31:54 GMT
Transfer-Encoding: chunked

2,,nfr06q05wg74ne5yr95kwsi1bm4fl0eq,2222,2020-07-21 03:11:55 +0600
22,ccj7b1oz91uc8vm7z5qjyctrjjpdb81g,nfr06q05wg74ne5yr95kwsi1bm4fl0eq,123,2020-07-21 03:16:38 +0600
23,ccj7b1oz91uc8vm7z5qjyctrjjpdb81g,nfr06q05wg74ne5yr95kwsi1bm4fl0eq,123,2020-07-21 03:16:42 +0600
24,ccj7b1oz91uc8vm7z5qjyctrjjpdb81g,nfr06q05wg74ne5yr95kwsi1bm4fl0eq,123,2020-07-21 03:16:43 +0600
25,ccj7b1oz91uc8vm7z5qjyctrjjpdb81g,nfr06q05wg74ne5yr95kwsi1bm4fl0eq,123,2020-07-21 03:16:45 +0600
26,ccj7b1oz91uc8vm7z5qjyctrjjpdb81g,nfr06q05wg74ne5yr95kwsi1bm4fl0eq,123,2020-07-21 03:16:46 +0600
```
