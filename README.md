# etcd

## Name

*redis* -  enables reading zone data from redis database

## Description

The *redis* plugin implements the same feature as bind.
 

 

## Syntax

~~~
redis [ZONES...]
~~~

* **ZONES** zones *redis* should be authoritative for.

 

~~~
etcd [ZONES...] {
    fallthrough [ZONES...]
    key_prefix KEY_PREFIX
    addresses ADDRESSES...
    username USERNAME
    password PASSWORD
    connect_timeout CONNECT_TIMEOUT
    read_timeout READ_TIMEOUT
    tls CERT KEY CACERT
}
~~~

 
* `tls` followed by:

    * no arguments, if the server certificate is signed by a system-installed CA and no client cert is needed
    * a single argument that is the CA PEM file, if the server cert is not signed by a system CA and no client cert is needed
    * two arguments - path to cert PEM file, the path to private key PEM file - if the server certificate is signed by a system-installed CA and a client certificate is needed
    * three arguments - path to cert PEM file, path to client private key PEM file, path to CA PEM
      file - if the server certificate is not signed by a system-installed CA and client certificate
      is needed.



## Examples

This is the default SkyDNS setup, with everything specified in full:

~~~ corefile
.:53 {
    log
    redis . {
      key_prefix coredns
      addresses 127.0.0.1:6379
      fallthrough
   }
   cache
   forward . 8.8.8.8:53 8.8.4.4:53
}
~~~

 

Multiple addresses are supported as well.

~~~
etcd skydns.local {
    addresses :6379 :6380
...
~~~

## zone format in redis db

### zones

Each record is stored in redis as a hash map with prefix `coredns`:*zone* as key
~~~
127.0.0.1:6379> KEYS *
 1) "coredns:net:example:mx"
 2) "coredns:arpa:in-addr:1:2:3:4"
 3) "coredns:net:example:txt"
 4) "coredns:net:example"
 5) "coredns:net:example:srv"
 6) "coredns:net:example:host"
 7) "coredns:net:example:ns"
 8) "coredns:net:example:host1"
~~~

~~~
127.0.0.1:6379> hgetall coredns:net:example
 1) "A"
 2) "[{\"ttl\":30,\"ip\":\"1.1.1.1\"}]"
 3) "SRV"
 4) "[{\"ttl\":10,\"priority\":10,\"weight\":1,\"port\":8080,\"target\":\"srv1.example.net\"},{\"ttl\":10,\"priority\":10,\"weight\":2,\"port\":8081,\"target\":\"srv2.example.net\"}]"
 5) "TXT"
 6) "[{\"ttl\":30,\"text\":\"hello word\"}]"
 7) "NS"
 8) "[{\"host\":\"ns1.example.net\"}]"
 9) "CAA"
10) "[{\"flag\":0,\"tag\":\"issue\",\"value\":\"dnspod.cn\"}]"
11) "SOA"
12) "{\"ns\":\"ns.dns.example.net\",\"Mbox\":\"hostmaster.example.net\",\"refresh\":86400,\"retry\":7200,\"expire\":3600,\"minTTL\":30}"
13) "MX"
14) "[{\"ttl\":10,\"host\":\"mail.example.net\",\"preference\":10}]"
~~~
*CNAME*
~~~
127.0.0.1:6379> hgetall  coredns:net:example:txt
1) "CNAME"
2) "[{\"ttl\":30,\"host\":\"example.net\"}]"
~~~

*PTR*
~~~
127.0.0.1:6379> hgetall coredns:arpa:in-addr:1:2:3:4
1) "PTR"
2) "[{\"host\":\"example.net\"}]"
~~~