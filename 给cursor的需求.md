
---
我现在需要你帮我使用go 用标准库实现一个 cli 交互的自动化工具
##### 让用户通过cli 输入账号,密码
将用户输入的账号密码 用一个变量: encode保存下来
账号%25%25%25密码%3D
账号和密码均使用base64加密,格式类似:
`MjAyMzEyMDA5Nzc4%25%25%25TGl1MDUwNDIw%3D`

##### 发送请求登陆拿到服务器返回的cookie
```txt
POST /ytkjxy_jsxsd/xk/LoginToXk HTTP/1.1
Host: jw.educationgroup.cn
Content-Length: 48

encoded=MjAyNDEyMDAzNzA3%25%25%25TXNoMDUwNjIx%3D
```
参照上面接口,返回的内容会是
```txt
HTTP/1.1 302 Found
Date: Wed, 25 Jun 2025 03:57:51 GMT
Content-Type: text/html
Content-Length: 0
Connection: keep-alive
Set-Cookie: HWWAFSESID=0fb737392385ab5141; path=/
Set-Cookie: HWWAFSESTIME=1750823871701; path=/
set-cookie: bzb_jsxsd=5F963D4A5CC5EF21780E5DDC73DF4D16; Path=/ytkjxy_jsxsd; HttpOnly
x-frame-options: SAMEORIGIN
pragma: No-cache
cache-control: no-cache
expires: Thu, 01 Jan 1970 00:00:00 GMT
location: https://jw.educationgroup.cn/ytkjxy_jsxsd/framework/xsrkxz.jsp
Set-Cookie: SERVERID=173; path=/
Set-Cookie: 961be5a6847249209aa003150e79ae88=WyI0Mjc0NzI4NjUwIl0; Expires=Wed, 25-Jun-25 04:17:51 GMT; Domain=jw.educationgroup.cn; Path=/; Secure; HttpOnly
Server: CW
```
你需要将所有的set cookie都存下来 在后续的请求中携带

##### 请求一个必要的认证接口
```txt
GET /ytkjxy_jsxsd/xsxk/xsxk_index?jx0502zbid=C260FE8330C34E8ABECB82E9ED5CE241 HTTP/1.1
Host: jw.educationgroup.cn
Cookie: bzb_jsxsd=5F963D4A5CC5EF21780E5DDC73DF4D16; HWWAFSESID=0fb737392385ab5141; HWWAFSESTIME=1750823871701; SERVERID=173; 961be5a6847249209aa003150e79ae88=WyI0Mjc0NzI4NjUwIl0

```

##### 获取课程列表 你只需要替换cookie即可,其他字段不变
```txt
POST /ytkjxy_jsxsd/xsxkkc/xsxkGgxxkxk?kcxx=&skls=&skxq=&skjc=&sfym=false&sfct=false&szjylb=&sfxx=true HTTP/1.1
Host: jw.educationgroup.cn
Cookie: bzb_jsxsd=5F963D4A5CC5EF21780E5DDC73DF4D16; HWWAFSESID=0fb737392385ab5141; HWWAFSESTIME=1750823871701; SERVERID=173; 961be5a6847249209aa003150e79ae88=WyI0Mjc0NzI4NjUwIl0
Content-Type: application/x-www-form-urlencoded; charset=UTF-8
Content-Length: 292

sEcho=1&iColumns=13&sColumns=&iDisplayStart=0&iDisplayLength=9999&mDataProp_0=kch&mDataProp_1=kcmc&mDataProp_2=xf&mDataProp_3=skls&mDataProp_4=sksj&mDataProp_5=skdd&mDataProp_6=xqmc&mDataProp_7=xxrs&mDataProp_8=xkrs&mDataProp_9=syrs&mDataProp_10=ctsm&mDataProp_11=szkcflmc&mDataProp_12=czOper
```

下面是响应体的内容,请你反序列化其中的一些字段
存入map中
```json
{"aaData":[{"kkdw":"0030","kcxzm":"14","szkcflmc":"人文科学（人文素养类）","parentjx0404id":null,"fj_filename":null,"xnxq01id":"2025-2026-1","zcxqjcList":[{"zc":"2","xq":"1","jc":"09"},{"zc":"2","xq":"1","jc":"10"},{"zc":"3","xq":"1","jc":"09"},{"zc":"3","xq":"1","jc":"10"},{"zc":"4","xq":"1","jc":"09"},{"zc":"4","xq":"1","jc":"10"},{"zc":"5","xq":"1","jc":"09"},{"zc":"5","xq":"1","jc":"10"},{"zc":"6","xq":"1","jc":"09"},{"zc":"6","xq":"1","jc":"10"},{"zc":"7","xq":"1","jc":"09"},{"zc":"7","xq":"1","jc":"10"},{"zc":"8","xq":"1","jc":"09"},{"zc":"8","xq":"1","jc":"10"},{"zc":"9","xq":"1","jc":"09"},{"zc":"9","xq":"1","jc":"10"},{"zc":"10","xq":"1","jc":"09"},{"zc":"10","xq":"1","jc":"10"},{"zc":"11","xq":"1","jc":"09"},{"zc":"11","xq":"1","jc":"10"},{"zc":"12","xq":"1","jc":"09"},{"zc":"12","xq":"1","jc":"10"},{"zc":"13","xq":"1","jc":"09"},{"zc":"13","xq":"1","jc":"10"},{"zc":"14","xq":"1","jc":"09"},{"zc":"14","xq":"1","jc":"10"},{"zc":"15","xq":"1","jc":"09"},{"zc":"15","xq":"1","jc":"10"},{"zc":"16","xq":"1","jc":"09"},{"zc":"16","xq":"1","jc":"10"},{"zc":"17","xq":"1","jc":"09"},{"zc":"17","xq":"1","jc":"10"}],"dwmc":"教务处（服务地方办公室）","ksfs":"9","pkrs":60,"xkrs":60,"ktmc":"临班809","tzdlb":"3","kcsx":"4","kch":"B0802504","syrs":"0","kxh":null,"fzmc":null,"cfbs":null,"sksj":"2-17周 星期一 9-10节","kcmc":"外国高等教育专题","kkapList":[{"jssj":"20:50","jzwmc":"第零教学楼","jgxm":"侯月华","skjcmc":"9-10","skzcList":["2","3","4","5","6","7","8","9","10","11","12","13","14","15","16","17"],"xq":"1","kbjcmsid":"3D61C36174274CA38474BFC85714EF1D","kkzc":"2-17","kssj":"19:10","jsmc":"虚拟教室_16","kkdlb":"1"}],"szkcfl":"1","dyrsbl":null,"ctsm":"","sklsid":"202212198","skls":"侯月华","kcjj":null,"xqid":"01","sfkfxk":"1","skdd":"虚拟教室_16","xf":2,"txsfkxq":"0","kcxzmc":"公共基础选修课","zxs":32,"jx0404id":"202520261000290","xqmc":"主校区","jxdg_filename":null,"jx02id":"00CB489842E148F7A7394DB3C829AD10","xxrs":60,"ggxxklb":"0","xbyq":null,"xbyqmc":null,"sftk":null},{"kkdw":"0030","kcxzm":"14","szkcflmc":"人文科学（人文素养类）","parentjx0404id":null,"fj_filename":null,"xnxq01id":"2025-2026-1","zcxqjcList":[{"zc":"2","xq":"1","jc":"09"},{"zc":"2","xq":"1","jc":"10"},{"zc":"3","xq":"1","jc":"09"},{"zc":"3","xq":"1","jc":"10"},{"zc":"4","xq":"1","jc":"09"},{"zc":"4","xq":"1","jc":"10"},{"zc":"5","xq":"1","jc":"09"},{"zc":"5","xq":"1","jc":"10"},{"zc":"6","xq":"1","jc":"09"},{"zc":"6","xq":"1","jc":"10"},{"zc":"7","xq":"1","jc":"09"},{"zc":"7","xq":"1","jc":"10"},{"zc":"8","xq":"1","jc":"09"},{"zc":"8","xq":"1","jc":"10"},{"zc":"9","xq":"1","jc":"09"},{"zc":"9","xq":"1","jc":"10"},{"zc":"10","xq":"1","jc":"09"},{"zc":"10","xq":"1","jc":"10"},{"zc":"11","xq":"1","jc":"09"},{"zc":"11","xq":"1","jc":"10"},{"zc":"12","xq":"1","jc":"09"},{"zc":"12","xq":"1","jc":"10"},{"zc":"13","xq":"1","jc":"09"},{"zc":"13","xq":"1","jc":"10"},{"zc":"14","xq":"1","jc":"09"},{"zc":"14","xq":"1","jc":"10"},{"zc":"15","xq":"1","jc":"09"},{"zc":"15","xq":"1","jc":"10"},{"zc":"16","xq":"1","jc":"09"},{"zc":"16","xq":"1","jc":"10"},{"zc":"17","xq":"1","jc":"09"},{"zc":"17","xq":"1","jc":"10"}],"dwmc":"教务处（服务地方办公室）","ksfs":"9","pkrs":60,"xkrs":60,"ktmc":"临班754","tzdlb":"3","kcsx":"4","kch":"B0802464","syrs":"0","kxh":null,"fzmc":null,"cfbs":null,"sksj":"2-17周 星期一 9-10节","kcmc":"品牌学","kkapList":[{"jssj":"20:50","jzwmc":"第零教学楼","jgxm":"胡君","skjcmc":"9-10","skzcList":["2","3","4","5","6","7","8","9","10","11","12","13","14","15","16","17"],"xq":"1","kbjcmsid":"3D61C36174274CA38474BFC85714EF1D","kkzc":"2-17","kssj":"19:10","jsmc":"虚拟教室_11","kkdlb":"1"}],"szkcfl":"1","dyrsbl":null,"ctsm":"","sklsid":"dfab48ed5aa94bd5ba5c1aea975afac1","skls":"胡君","kcjj":null,"xqid":"01","sfkfxk":"1","skdd":"虚拟教室_11","xf":2,"txsfkxq":"1","kcxzmc":"公共基础选修课","zxs":32,"jx0404id":"202520261000235","xqmc":"主校区","jxdg_filename":null,"jx02id":"01FA6AA82C034F5E94080CA1F1B714A2","xxrs":60,"ggxxklb":"0","xbyq":null,"xbyqmc":null,"sftk":null}],"sEcho":"1","iTotalRecords":260,"iTotalDisplayRecords":260,"jfViewStr":""}
```

将  "kch": "B0802464", 的B0802464 作为key,
 "jx0404id": "202520261000235",的  202520261000235 作为value
 存入map中


##### 用户选课
CLI将 kch 打印出来,
让用户输入要选哪些kch,存入slice中,可多选
直到输入某个值作为停止

选课接口:
```txt
GET /ytkjxy_jsxsd/xsxkkc/ggxxkxkOper?cfbs=null&jx0404id=202520261000190&xkzy=&trjf=&_=1750820826702 HTTP/1.1
Host: jw.educationgroup.cn
Cookie: bzb_jsxsd=8EC1444C9D29159D81999A14DBA58E38; HWWAFSESID=e564473d20fcc50393; HWWAFSESTIME=1750820620392; SERVERID=174; 961be5a6847249209aa003150e79ae88=WyI0Mjc0NzI4NjUwIl0


```
将jx0404id替换为用户选择的kch的map对应的value,即可完成选课

响应内容
```txt
`{"success":true,"message":"选课成功","jfViewStr":""}
```

具体实现:
遍历slice,使用线程间通信的方式,为每一个选课都开一个协程,
协程1s请求一次接口
直到返回内容的message 为 选课成功 该协程完成任务,否则打印message内容,

某一个协程完成抢课,则打印日志告诉某课程号抢课成功,退出该协程

当所有协程正常退出后,停止程序,
打印选课成功的课程号

