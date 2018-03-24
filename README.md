# btcd2sql 

Convertor from btcsuit/btcd blockchain database to sql format.

Convertor based on the https://github.com/webdeveloppro/cryptopiggy app
For database structure you can take a look on the https://github.com/webdeveloppro/cryptopiggy/sql

Its pretty simple idea actually:
- Start from genesic block
- Move one by one, log block data, transfer data
- Update hash addresses during changes

Challange might be with a list of the "bad" txin/txout data, like:

Non-zero outputs which address is unknows:
  tx_hash=b728387a3cf1dfcff1eef13706816327907f79f9366a7098ee48fc0c00ad2726,
  tx_hash=9740e7d646f5278603c04706a366716e5e87212c57395e0d24761c0ae784b2c6,
  tx_hash=9969603dca74d14d29d1d5f56b94c7872551607f8c2d6837ab9715c60721b50e,
  tx_hash=b8fd633e7713a43d5ac87266adc78444669b987a56b3a65fb92d58c2c4b0e84d,
  tx_hash=60a20bd93aa49ab4b28d514ec10b06e1829ce6818ec06cd3aabd013ebcdc4bb1,
  tx_hash=f003f0c1193019db2497a675fd05d9f2edddf9b67c59e677c48d3dbd4ed5f00b,
  tx_hash=fa735229f650a8a12bcf2f14cca5a8593513f0aabc52f8687ee148c9f9ab6665,
  tx_hash=b38bb421d9a54c58ea331c4b4823dd498f1e42e25ac96d3db643308fcc70503e,
  tx_hash=9c08a4d78931342b37fd5f72900fb9983087e6f46c4a097d8a1f52c74e28eaf6,
  tx_hash=c0b69d1e5ed13732dbd704604f7c08bc96549cc556c464aa42cc7525b3897987,
  tx_hash=aea682d68a3ea5e3583e088dcbd699a5d44d4b083f02ad0aaf2598fe1fa4dfd4,
 
Second challange is amount of data and insert/update race competitions during data import.
Migrating to jsonb for tx-inputs and tx-outputs speed up application for and 10 times

## Install process

You need to have blocks_ffldb folder ready with blockchain data and webdeveloppro/cryptopiggy repositary

```
git clone https://github.com/webdeveloppro/btcd2sql 
cd btcd2sql/leveldb2sql
vi .env.bash.sh  <- edit your database credentials
go build
go run
```

this will start migration process. Hole process will take more ~10 hours (for 150GB of blockchain data)
