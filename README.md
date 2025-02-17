<!--
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2025-02-16 22:23:16
 * @LastEditTime: 2025-02-17 22:11:34
 * @LastEditors: FunctionSir
 * @Description: README.md
 * @FilePath: /wol-http/README.md
-->
# wol-http

Wake on LAN plus HTTP.  
Current version: 0.1.0 (SatenRuiko).  

## How to config

Please config it base on the example.conf.  
Note: the token required at least 8 chars!  
P.S. If you don't want HTTPS, you can just don't set "Cert" and "Key", but if you want HTTPS, BOTH of them are required.  

## How to use

Send GET requests.  
URL: http(s)://example.org:port/token/action/by-what/key  
For example:  
<https://example.org:2420/some-token/info/name/server_a>  
<https://example.org:2420/some-token/wake/alias/sa>  
<https://example.org:2420/some-token/info/ip/114.5.1.4>  
<https://example.org:2420/some-token/wake/mac/14:51:41:14:51:41>  
You can specify the target by name/alias/IP/MAC.  
Use action "info" to get the information you set.  
Use action "wake" to try to wake it up on LAN.  
If succeed, the response will have a status code 200 and a body of the command output (stdout and stderr combined output).  
