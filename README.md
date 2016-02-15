# rmlp

rmlp is Redis Monitor Log Profiler.

# Installation

# Useage

Read from stdin or an input file (-f)

```
redis-cli monitor | rmlp

rmlp -f redis-monitor.log
```

```
$ rmlp -f redis-monitor-example.log

Overall Stats
==================================
LineCount        15000


Commands Rate
==================================
LRANGE   4000
PING     2000
LPUSH    2000
INCR     1000
SADD     1000
SPOP     1000
MSET     1000
LPOP     1000
SET      1000
GET      1000

Heavy Commands
==================================
Command Sum(msec)
LRANGE   0.803558
PING     0.106445
LPUSH    0.099761
MSET     0.092236
SET      0.060840
INCR     0.056400
SADD     0.053511
GET      0.051652
LPOP     0.047541
SPOP     0.038231

Slowest Calls
==================================
KEY                             Count   Max(msec)        Avg(msec)
PING                             2000    0.001511        0.000053
LRANGE mylist                    4000    0.001358        0.000201
MSET key:__rand_int__            1000    0.000959        0.000092
LPUSH mylist                     2000    0.000856        0.000050
INCR counter:__rand_int__        1000    0.000837        0.000056
SET key:__rand_int__             1000    0.000806        0.000061
SPOP myset                       1000    0.000705        0.000038
SADD myset                       1000    0.000538        0.000054
GET key:__rand_int__             1000    0.000480        0.000052
LPOP mylist                      1000    0.000427        0.000048

$ rmlp -f redis-monitor-example.log -s avg
Overall Stats
==================================
LineCount        15000


Commands Rate
==================================
LRANGE   4000
PING     2000
LPUSH    2000
INCR     1000
SADD     1000
SPOP     1000
MSET     1000
LPOP     1000
SET      1000
GET      1000

Heavy Commands
==================================
Command Sum(msec)
LRANGE   0.803558
PING     0.106445
LPUSH    0.099761
MSET     0.092236
SET      0.060840
INCR     0.056400
SADD     0.053511
GET      0.051652
LPOP     0.047541
SPOP     0.038231

Slowest Calls
==================================
KEY                             Count   Max(msec)        Avg(msec)
LRANGE mylist                    4000    0.001358        0.000201
MSET key:__rand_int__            1000    0.000959        0.000092
SET key:__rand_int__             1000    0.000806        0.000061
INCR counter:__rand_int__        1000    0.000837        0.000056
SADD myset                       1000    0.000538        0.000054
PING                             2000    0.001511        0.000053
GET key:__rand_int__             1000    0.000480        0.000052
LPUSH mylist                     2000    0.000856        0.000050
LPOP mylist                      1000    0.000427        0.000048
SPOP myset                       1000    0.000705        0.000038
```
