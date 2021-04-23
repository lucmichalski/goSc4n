We are **gosc4n** team, students of FPT University <br/> <br/>
<p align="center">
  <img alt="gosc4n" src="https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/Logo_goSc4n.png" height="140"/>
  <p align="center">
  </p>
</p>

**gosc4n** is a powerful, flexible and easily extensible framework written in Go for building your own Web Application
Scanner.

<p align="center">
<img alt="gosc4n" src="https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/Roadmap.png" height="500" />
</p>

## Painless integrate gosc4n into your recon workflow?

<p align="center">
  <img alt="paramSpider" src="https://raw.githubusercontent.com/devanshbatham/ParamSpider/master/static/banner.PNG" height="200" />
  <p align="center"><br /></p>
   <p align="center">
    <img alt="spider" src="https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/paramspider.PNG" />
  </p> 
</p>

## Installation

Download [precompiled version here](https://github.com/goSc4n/goSc4n/releases).

If you have a Go environment, make sure you have **Go >= 1.13** with Go Modules enable and run the following command.


## Usage
![Architecture](https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/scanusage1.png)

![Architecture](https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/scanusage2.PNG)

# Scan Usage example:
![Architecture](https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/scanexample.PNG)


 
# Fuzz Usage:
![Architecture](https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/fuzzusage.png)


```shell
# Fuzz Usage example:
  
  fuzz --quite --site "https://google.com/"
  fuzz --site "https://google.com/" --output ouput --concurrent 10 --depth 10
  fuzz --sites sites.txt --outpud output --concurrent 10 --depth 1
  fuzz --sites sites.txt --outpud output --concurrent 10 --depth 1 --threads 20
```
 
# Spider Usage:
![Architecture](https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/spiderusage.png)


```shell
# Spider Usage example:
  
  spider --domain hackerone.com
  spider --domain hackerone.com --level high
  spider --domain hackerone.com --exclude php,jpg --output hackerone.txt
  spider --domain hackerone.com --quiet
```






### HTML Report summary

![Architecture](https://raw.githubusercontent.com/goSc4n/goSc4n/main/picture/summary.png)



### Planned Features

* Adding more signatures.
* Adding more input sources.
* Adding more APIs to get access to more properties of the request.
* Adding more action on Web UI.
* Integrate with many other tools.

## Contribute

If you have some new idea about this project, issue, feedback or found some valuable tool feel free to open an issue for
just DM me via @gosc4n. Feel free to submit new signature to
this [repo](https://github.com/goSc4n/goSc4n/tree/main/base-signatures).





## License

`gosc4n` is made with ♥ by [gosc4nTeam]) and it is released under the MIT license.

