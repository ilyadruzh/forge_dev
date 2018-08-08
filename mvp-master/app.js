'use strict';

var log4js = require('log4js');
var logger = log4js.getLogger('Loan');
var express = require('express');
var bodyParser = require('body-parser');
var http = require('http');
var app = express();
var config = require('./config.json');
require('./config.js');
var hfc = require('fabric-client');

var host = process.env.HOST || hfc.getConfigSetting('host');
var port = process.env.PORT || hfc.getConfigSetting('port');
if (host.indexOf("http://")==-1 && host.indexOf("https://")==-1){
    host = "http://"+host;
}
var httpAddr = host;
if(port != 80){
    httpAddr = httpAddr+":"+port;
}


///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// SET CONFIGURATONS ////////////////////////////
///////////////////////////////////////////////////////////////////////////////

app.use(bodyParser.json());
app.use(bodyParser.urlencoded({
	extended: false
}));

app.set('secret', config.jwt_secret);









///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// START SERVER /////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
var server = http.createServer(app).listen(port, function() {});
logger.info('****************** SERVER STARTED ************************');
logger.info('**************  ' + httpAddr +	'  ******************');
server.timeout = 240000;


module.exports.app = app;
module.exports.port = port;
module.exports.host = host;
require('./app/rest');


