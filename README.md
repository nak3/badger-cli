# badger-cli

A simple interactive and batch shell for [badger key-value store](https://github.com/dgraph-io/badger).

### Build

~~~
git clone https://github.com/nak3/badger-cli.git
cd badger-cli
go build ./...
~~~

### Usage

~~~
$ ./badger-cli -h
Usage:

  ./badger-cli [OPTIONS] [SYNTAX]

Syntax:
  get <KEY>                 get value by key
  set <KEY> <VALUE> <META>  set item by key and value
  delete <KEY>              delete item by key
  dump                      dump item list
  (nil)                     no syntax starts interactive shell
Options:
  -dir string
    	The Badger database's index directory
  -value-dir string
    	The Badger database's value log directory, if different from the index directory
~~~

### Example

#### interactive mode

~~~
mkdir -p /tmp/badger-test
~~~

1) start interactive mode

~~~
$ ./badger-cli --dir=/tmp/badger-test
/tmp/badger-test >
~~~

2) set item by key and value 

~~~
/tmp/badger-test > set apple 100
/tmp/badger-test > set orange 200
/tmp/badger-test > set grape 300
~~~

3) get item by key

~~~
/tmp/badger-test > get apple
100
~~~

4) dump key value list

~~~
/tmp/badger-test > dump
apple 100
grape 300
orange 200
~~~

5) Exit interactive mode

~~~
/tmp/badger-test > exit (\q, Ctr+D or Ctr+C)
~~~


#### batch mode

1) set item by key and value 

~~~
./badger-cli --dir=/tmp/badger-test set banana 500
~~~

2) get item by key

~~~
./badger-cli --dir=/tmp/badger-test get banana 
500
~~~

3)  dump key value list

~~~
./badger-cli --dir=/tmp/badger-test dump
banana 500
~~~
