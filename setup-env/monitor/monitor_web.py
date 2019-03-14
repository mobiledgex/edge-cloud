#!/usr/bin/env python
#
##################################################################
#
import web
from web import form
import os
import glob
import traceback
import sys
import time
import glob
import json  
import string 
import re
import urllib2
from datetime import datetime
import hashlib
import subprocess
from multiprocessing.dummy import Pool as ThreadPool
import base64
import importlib
from yaml import load, dump
try:
    from yaml import CLoader as Loader, CDumper as Dumper
except ImportError:
    from yaml import Loader, Dumper

Urls = (
     '/login', 'login',
     '/logout', 'logout',
     '/status', 'show_status'
)


MouseOver = "onmouseover=\"this.style.backgroundColor=\'#00ff00\';\" onmouseout=\"this.style.backgroundColor=\'#d4e3e5\';\""

def getAuth():
     readWrite = ("mexadmin:mexadmin")
     
     readOnly = ("mexreadonly:mexreadonly")
     
     auth = web.ctx.env.get('HTTP_AUTHORIZATION')
     authreq = False
     username = None
     password = None
     if auth is None:
         return "NOAUTH"
     else:
        auth = re.sub('^Basic ','',auth)
        username,password = base64.decodestring(auth).split(':')
        userpass = username+":"+password
        if userpass in readWrite:
           return "READWRITE"
        if userpass in readOnly:
           return "READONLY"

     return "NOAUTH"   

class logout:
   def GET(self):
        web.ctx.status = '401 Unauthorized'

        html = "<html><br>Logged Out<br><br>"
        homeUrl = web.ctx["home"]+"/status"
        homeLink= "<a href=%s>%s</a>" % (homeUrl,"Back to Demo Status")
        html += homeLink

        html += "</html>"
        return html
       
class login:
    def GET(self):    
       authRc = getAuth()

       if authRc == "NOAUTH":
           web.header('WWW-Authenticate','Basic realm="monitor authenticate"')
           web.ctx.status = '401 Unauthorized'
           return
       raise web.seeother('/status')
         

class show_status:

   def checkDmeHealth(self, endpoint, appname, devname):
       type,name,uri = endpoint.split("|")
       p = subprocess.Popen(["/usr/local/bin/edgectl --addr "+uri+" --tls /root/tls/mex-client.crt dme RegisterClient --appname \""+appname+"\" --appvers \"1.0\" --devname \""+devname+"\""], stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
       out,err = p.communicate()
       print ("DME OUT %s" % out)
       if "RS_SUCCESS" in out:
           return type,name,uri,"OK"
       else:
           return type,name,uri,"FAIL"


   def checkHealth(self,endpoint):
       print ("checkhealth for endpoint: %s\n" % endpoint)
       type,name,uri = endpoint.split("|")
       if "DME" in name:
           # see if we can register with the sample app
           return self.checkDmeHealth(endpoint, "MobiledgeX SDK Demo", "MobiledgeX")
       try:
         headers = {}
         msg = ""
         req = urllib2.Request("http://"+uri, msg, headers)
         print ("posting to %s\n" % uri)
         response = urllib2.urlopen(req, timeout=3)
         return type,name,uri,"OK"
       except urllib2.HTTPError, e:
          if (e.code == 404) and ("facedetection" in uri):
              ## this is ok, we need a health check url
              return type,name,uri,"OK"
          return type,name,uri,"FAIL - %s %s" % (e.code,e.reason)
       except urllib2.URLError as ue:
          return type,name,uri,"FAIL - %s" % (ue.reason)
       except Exception as e2:
          print("unknown on post to url: %s -- %s" % (uri,e2))
          return type,name,uri,"FAIL - %s" % e2

   def getUrisForApp(self, appname):
       ## todo: configurable endpoints
       p = subprocess.Popen(["/usr/local/bin/edgectl --addr mexdemo.ctrl.mobiledgex.net:55001 --tls /root/tls/mex-client.crt controller ShowAppInst --key-appkey-name \""+appname+"\""], stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
       
       uris = []
       out,err = p.communicate()
       data = load(out, Loader=Loader)
       print ("YAMLDATA %s" % data)
       if not data or len(data) == 0:
          print "ERROR: no data\n"
          return uris
        
       for appinst in data: 
          pubport = appinst['mappedports'][0]['publicport']
          fqdnprefix = appinst['mappedports'][0]['fqdnprefix']
          uri  = appinst['uri']
      
          fulluri = fqdnprefix+uri+":"+str(pubport)
          print "fulluri: %s " % fulluri
          uris.append(fulluri)
       return uris
                  
   def getHealthResults(self, appnames):
        results = dict()
        pool = ThreadPool(20)

        itemsToCheck = []
        
        for appname in appnames:
           try:
             uris = self.getUrisForApp(appname)
             for uri in uris:
               itemsToCheck.append("App Instances|"+appname+"|"+uri) 
           except Exception as e:
              print "Exception getting app uris for %s - %s " % (appname, e)
        ## todo: remove hardcoded endpoints
        gws = ("Token Simulator|mexdemo.tok.mobiledgex.net:9999", 
               "Location API Simulator|mexdemo.locsim.mobiledgex.net:8888",
               "DME - MWC|gddt.dme.mobiledgex.net:50051", 
               "DME - Demo|mexdemo.dme.mobiledgex.net:50051")
        
        for gw in gws:
           gwname,uri = gw.split("|")
           itemsToCheck.append("API Gateways|"+gwname+"|"+uri)

        checkresults = pool.map(self.checkHealth,itemsToCheck)
        print ("checkresults out %s" % checkresults)
        for item in checkresults:
               results[item[0]+"|"+item[1]+"|"+item[2]] = item[3]

        return results

   def __init__(self):
     print("show_status INIT method\n")
     self.style = None

   def readCss(self):
      home = os.environ['HOME']
      with open(home+"/monitor.css", 'r') as myfile:
           self.style = myfile.read()
      myfile.close()

   def printHtmlTableHeader(self, columns, tableclass):
     rc = ""
     numcols = len(columns)
     rc +=  "<table class=\""+tableclass+"\"><tr>"
     for col in columns: 
       rc += "<th>"+col+"</th>"
     rc += "\n"
     return rc 
   
   def printHtmlTableRow(self, columns, colorCode):
       color = ""  
       rc = ""
       if colorCode:
          color = ";color:"+colorCode+";font-weight:bold"     

       numcols = len(columns) 
 
       rc +=  "<tr "+MouseOver+">"

       colsprinted = 0
       first = True
         
       for col in columns:
         colsprinted += 1

         if first:
           rc += "<td style=\"width:150px\">"+str(col)+"</td>\n" 
         else:

           colorUsed = ""
           if colsprinted == numcols:
               colorUsed = color
               
           rc += "<td style =\"width:80px "+colorUsed+" \" >"+str(col)+"</td>\n"
         first = False
 
       rc += "</tr>\n"

       return rc
 
              
   def GET(self): 

     if web.ctx.env.get('HTTP_AUTHORIZATION') is None:
             raise web.seeother('/login')
     try:
        self.readCss()
        html = "<html><head><title>System Status</title><style type = \"text/css\">\n"
        html += self.style+"\n</style></head><body>"

        html += "</style></head></body></head>\n"
        html += "<p><font size=\"5\" color=\"#003300\">MEX Demo System Status</font></p>\n" 
        html += "<br><br>\n" 
        hdrcols = ["Service Type", "Name", "URI", "Status"]

        html += self.printHtmlTableHeader(hdrcols,"hovertable")

        appnames = ("MobiledgeX SDK Demo", "Face Detection Demo")
        results = self.getHealthResults(appnames)

        for res in sorted(results.iterkeys()):
           status = results[res]
           (type,name,uri) = res.split("|")

           color = "red"
           if status == "OK":
              color = "green"

           html += self.printHtmlTableRow((type,name,uri,status), color)
           html += "</tr>\n"
        html += "</table><br><br>"
        home = web.ctx["home"]
        dispUrl = home+"/logout"
        reflink = "<p><a href=%s>%s</a></p>" % (dispUrl, "Logout")
        html += reflink
        html += "</html>"
        return html
          
     except Exception as e:
         exc_type, exc_value, exc_traceback = sys.exc_info()
         print("Unexpected Exception  %s Traceback %s" % (e,traceback.format_tb(exc_traceback)))
         raise web.internalerror("error starting app "+str(e))


if __name__ == "__main__":
    try:
       app = web.application(Urls, globals())
       app.run()
    except Exception as e:
        exc_type, exc_value, exc_traceback = sys.exc_info()
        raise web.internalerror("error starting app "+str(e))
