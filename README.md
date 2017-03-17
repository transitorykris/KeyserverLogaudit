# Upspin Logaudit

A service that checks and reports on the integrity of key.upspin.io/log

* `api/` contains the front end for the service
* `worker/` is where the auditing is actually done
* `db/` contains [goose](https://bitbucket.org/liamstask/goose/) migrations

## Heads up

This is opened up for auditing and isn't yet in a state that can easily be run locally. I had to tease this apart from 
code that I cannot open source. I will be tidying this in the future so it can stand on its own.

Pull requests happily accepted!
