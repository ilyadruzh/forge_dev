var log4js = require('log4js');
const fs = require("fs");
var path = require('path');
var logger = log4js.getLogger('api');
logger.setLevel('INFO');
var helper = require('./helper.js');
var bapi = require('./brainy_api.js');
var invoke = require('./invoke-transaction.js');
var query = require('./query.js');
var config = require('../config.json');
var org = config.myOrganisation;
var robot = require('./robot.js');
var dateFormat = require('dateformat');
var PathOfBackup={}

var createBackupPath = function() {
    PathOfBackup.date = Date.now()
    PathOfBackup.dateStr = dateFormat(PathOfBackup.date, "yyyy-mm-dd");
    PathOfBackup.invoke = path.join(__dirname, '..', 'backups', 'invoke_' + PathOfBackup.dateStr + '.dat');
    PathOfBackup.query = path.join(__dirname, '..', 'backups', 'query_' + PathOfBackup.dateStr + '.dat');

    var tmp = PathOfBackup.dateStr.split("-");
    var timeToReset = new Date(parseInt(tmp[0]),parseInt(tmp[1])-1,parseInt(tmp[2]),23,59,59).getTime() //Date.parse(PathOfBackup.dateStr)//,"yyyy-MM-dd")
    timeToReset = timeToReset-PathOfBackup.date;

    setTimeout(function () {
        createBackupPath();
    },timeToReset)
}

createBackupPath();


var setEthToUserInBc = function(login,wallet,eth,callback){
    //"rom","wallet","eth"
    invoking([login,wallet,eth], "setUserEth", function(result){
        callback(result);
    });
}
var getOperations = function(login,filter,periodStart,periodFinish,callback){
    // 0 - login
    //      FILTER
    // 1 - filter (if filter=="" then return all)
    // 2 - periodStart (if periodStart = "" return all)
    // 3 - periodEnd   (if periodEnd   = "" return time.now())
    querying([login,filter,periodStart,periodFinish], "getOperationData", function(res){
        callback(res);
    });
}

var createSchedule = function(login,orderId,callback){
    // tranches - array of tranche JSON
    // 0-login,1-OrderId,2-scheduleId,3-amount,4-chargeIssueFee,5-issued,6-activeBefore,7-tranches
    getOrder(login,orderId,function (order) {
        var date = dateFormat(Date.now(), "yyyy-mm-dd");
        bapi.getPaymentsSchedule(order.sum,order.percent, date,order.period+"", function(schedule){
            if(schedule && schedule.status == "ok") {
                invoking([login, orderId, "", "" + schedule.data.amount, "" + schedule.data.chargeIssueFee, "" + schedule.data.issued,
                    "" + schedule.data.activeBefore, JSON.stringify(schedule.data.tranches)], "createSchedule", function (res) {

                    getTrans(res.trxnId, login, function (result2) {
                        try {
                            var action = result2.transactionEnvelope.payload.data.actions[0];
                            var proposal_response_payload = action.payload.action.proposal_response_payload;
                            var payload = proposal_response_payload.extension.response.payload;
                            var data;
                            if (payload) {
                                data = JSON.parse(payload);
                                res.scheduleId = ""+data.scheduleId;
                                callback(res);
                            }
                        } catch (e) {
                            console.log(e)
                            callback({"status": false, "description": e.message + ""});
                        }
                    })
                });
            } else {
                if(schedule && schedule.data && schedule.data.code) {
                    console.log("ERROR: " + schedule.data.code);
                    callback({"status": false, "description": schedule.data.code + ""});
                } else {
                    console.log("ERROR: ");
                    callback({"status": false, "description": "ERROR"});
                }
            }
        })
    })
}

var setActiveSchedule = function(login,scheduleId,orderId,callback){
    // 0 - Login,
    // 1 - ScheduleId,
    // 2 - OrderId
    invoking([login,scheduleId,orderId], "setActiveSchedule", function(result){
        callback(result);
    });
}

var getScheduleById = function(login,scheduleId,callback){
    // 0 - Login, 1 - scheduleId
    querying([login,scheduleId], "getScheduleById", function(result){
        callback(result);
    });
}

var getScheduleList = function(login,orderId,callback){
    // 0 - Login, 1 - OrderId
    querying([login,orderId], "getScheduleList", function(result){
        callback(result);
    });
}

var getListOfTranchesByDate = function(login,dateStart,dateFinish,callback){
    // 0 - Login,
    // 1 - dateStart  format '2017-06-02'
    // 2 - dateFinish format '2017-07-02'
    querying([login,dateStart,dateFinish], "getListOfTranchesByDate", function(result){
        callback(result);
    });
}

var getActiveSchedule = function(login,orderId,callback){
    // 0 - Login, 1 - OrderId
    querying([login,orderId], "getActiveSchedule", function(result){
        callback(result);
    });
}


var tokenTransferInBc = function(login,tx,from,to,amount,orderId,callback){
    //"Login - 0, trHash-1, from - 2, to - 3, amount - 4", orderId
    invoking([login,tx,from,to,amount,orderId], "tokenTransfer", function(res){
        callback(res);
    });
}

var getHistoryTransferToken = function(login,eth,periodStart,periodFinish,callback){
    // 0 - login
    //      FILTER
    // 1 - eth (if eth=="" then return all)
    // 2 - periodStart (if periodStart = "" return all)
    // 3 - periodEnd   (if periodEnd   = "" return time.now())
    querying([login,eth,periodStart,periodFinish], "getHistoryTransferToken", function(res){
        callback(res);
    });
}

var createOrderInBc = function(login,period,borrower,lender,sum,percent,startDate,finishDate,callback){
    //Login - 0, period, Borrower, Lender, Sum, Percent, StartDate, FinishDate
    invoking([login,period,borrower,lender,sum,"-1",startDate,finishDate], "createOrder", function(res){

        getTrans(res.trxnId, login, function(result2){
            try {
                var action = result2.transactionEnvelope.payload.data.actions[0];
                var proposal_response_payload = action.payload.action.proposal_response_payload;
                var payload = proposal_response_payload.extension.response.payload;
                var data;
                if(payload) {
                    data = JSON.parse(payload);
                    res.orderId = data.orderId;
                    getUser(login, function(user){
                        bapi.getPercentage(user.brainyId, sum, function(perc,contractId){
                            if(contractId=="") contractId = "0";
                            updateOrder(login,period,borrower,lender,sum,""+perc,startDate,finishDate,""+data.orderId,""+contractId,function(result){
                                logger.info(JSON.stringify(result))
                                logger.info("[orderId="+data.orderId+",user="+login+"] Percent was updated successfully")
                            })
                        })
                    })
                    callback(res);
                }
            } catch(e){
                console.log(e)
                callback({"status":false,"description":e.message+""});
            }

        })
    });
}

var createUserInBc = function(login,pass,firstName,lastName,patronymic,birthDate,passport,email,mobilePhone,
                              role,props,wallet,startBalance,gender,inn,bureauConsentDate,addressData,
                              registrationAddressData,managerId,eth,
                              callback){

    // 0  - login
    // 1  - pass
    // 2  - firstName
    // 3  - lastName
    // 4  - patronymic
    // 5  - birthDate
    // 6  - passport
    // 7  - email
    // 8  - mobilePhone
    // 9  - role
    // 10 - props
    // 11 - wallet
    // 12 - balance   - float64
    // 13 - gender    - uint32
    // 14 - inn
    // 15 - bureauConsentDate
    // 16 - addressData
    // 17 - registrationAddressData
    // 18 - managerId  - uint64
    // 19 - eth
    // 20 - brainyId

    bapi.createPerson(firstName,lastName,patronymic,birthDate,passport,email,mobilePhone,
        addressData,bureauConsentDate,registrationAddressData, function(personData){
            invoking([login,pass,firstName,lastName,patronymic,birthDate,passport,email,mobilePhone,role,props,wallet,startBalance,
                gender,inn,bureauConsentDate,addressData,registrationAddressData,managerId,eth,""+personData.id], "registerUser", function(res){
                callback(res);
            });
        }
    );
}

var changeUserPasswordInBc = function(login,odPass,newPass,callback){
    //"ivan","123","1234"
    invoking([login,odPass,newPass], "changeUserPassword", function(res){
        callback(res);
    });
}

var confirmOrder = function(login,orderId,callback){
    //"ivan","123123"
    invoking([login,orderId], "confirmOrder", function(res){
        callback(res);
    });
}

var repaymentLoan = function(login,orderId,sum,callback){
    //"ivan","123123","10000"
    invoking([login,orderId,sum], "repaymentLoan", function(res){
        callback(res);
    });
}

var activateUserInBc = function(login,pass,callback){
    //"ivan","123"
    invoking([login,pass], "activateUser", function(res){
        callback(res);
    });
}

var activateOrderInBc = function(login,orderId,callback){
    //"ivan","123"
    invoking([login,orderId], "activateOrder", function(res){
        callback(res);
    });
}

var activateOffer = function(login,orderId,callback){
    //"ivan","123"
    invoking([login,orderId], "activateOffer", function(res){
        callback(res);
    });
}


var putOther = function(login,key,value,callback){
    invoking([login,key,value], "putOther", function(res){
        callback(res);
    });
}

var getOther = function(login,key,callback){
    querying([login,key], "getOther", function(res){
        callback(res);
    });
}

var userLogin = function(login,pass,callback){
    //"ivan","123"
    querying([login,pass], "userLogin", function(res){
        callback(res);
    });
}


var getSumToRepayment = function(login,orderId,callback){
    //"ivan", "123123"
    querying([login,orderId], "getSumToRepayment", function(res){
        callback(res);
    });
}

var closeOrder = function(login,orderId,callback){
    //"ivan","123"
    invoking([login,orderId], "closeOrder", function(res){
        callback(res);
    });
}

var cancelOrder = function(login,orderId,callback){
    //"ivan","123"
    invoking([login,orderId], "cancelOrder", function(res){
        callback(res);
    });
}

var getBorrows = function(login,borrowerId,callback){
    //"ivan", "123123"
    querying([login,borrowerId], "getBorrows", function(res){
        callback(res);
    });
}

var getLends = function(login,lenderId,callback){
    //"ivan", "123123"
    querying([login,lenderId], "getLends", function(res){
        callback(res);
    });
}

var getUser = function(login,callback){
    //"ivan"
    querying([login], "getUser", function(res){
        delete res.orders;
        delete res.schedules;
        callback(res);
    });
}

var getOrder = function(login, orderId, callback){
    //"ivan", orderId
    querying([login, orderId], "getOrder", function(res){
        callback(res);
    });
}

var getOffer = function(login, offerId, callback){
    //"ivan", offerId
    querying([login, offerId], "getOffer", function(res){
        callback(res);
    });
}

var getUserTranches = function(login,dateStart,dateFinish,callback){
    // 0 - Login,
    // 1 - dateStart format '2017-06-02'
    // 2 - dateFinish format '2017-06-02'
    querying([login, dateStart, dateFinish], "getUserTranches", function(res){
        callback(res);
    });
}

var getBorrowOffers = function(login, sumSt, percentSt, periodSt, sumFin, percentFin, periodFin, callback){
    //"ivan", sumSt, percentSt, periodSt, sumFin, percentFin, periodFin
    querying([login, sumSt, percentSt, periodSt, sumFin, percentFin, periodFin], "getBorrowOffers", function(res){
        callback(res);
    });
}

var getLendOffers = function(login, sumSt, percentSt, periodSt, sumFin, percentFin, periodFin, callback){
    //"ivan", sumSt, percentSt, periodSt, sumFin, percentFin, periodFin
    querying([login, sumSt, percentSt, periodSt, sumFin, percentFin, periodFin], "getLendOffers", function(res){
        callback(res);
    });
}

var updateOrder = function(login,period,borrower,lender,sum,percent,startDate,finishDate,orderId,contractId, callback){
    //{Login - 0, period, Borrower, Lender, Sum, Percent, StartDate, FinishDate, orderId}

    getOrder(login, orderId, function(order){
        if(order && order.status) {
            if(order.status == "ACTIVE" || order.status == "CLOSED" ){
                callback({"login":login,"status":false, "description":"Cannot update order, because current status='"+order.status+"'"});
            } else {
                var period2 = tryVsData(period, order.period + "");
                var borrower2 = tryVsData(borrower, order.borrower);
                var lender2 = tryVsData(lender, order.lender);
                var sum2 = tryVsData(sum, order.sum + "");
                var percent2 = tryVsData(percent, order.percent + "");
                var startDate2 = tryVsData(startDate, order.startDate);
                var finishDate2 = tryVsData(finishDate, order.finishDate);
                var contractId2 = tryVsData(contractId, order.contractId + "");

                invoking([login, period2, borrower2, lender2, sum2, percent2, startDate2, finishDate2, orderId, contractId2], "updateOrder", function (res) {
                    callback(res);
                });
            }
        } else callback(order);
    })
}

var paymentTranche = function(login,dateOfTranche,scheduleId,eachRepaymentFee, callback){
    // 0 - Login,
    // 1 - dateOfTranche  format '2017-06-02'
    // 2 - scheduleId
    // 3 - eachRepaymentFee
    invoking([login,dateOfTranche,scheduleId,eachRepaymentFee], "paymentTranche", function(res){
        callback(res);
    });
}


//
// var getUserOffers = function(login, callback){
//     //"ivan"
//     querying([login], "getUserOffers", function(res){
//         callback(res);
//     });
// }


/////////////////////////////////////////////////////////////////////////////////

var invoking = function(args, functionName, callback) {

    invoke.invokeChaincode(getAllPeers(), config.channelName, config.chaincodeName, functionName, args, args[0], org)
        .then(function (message) {
            if (message && message !== 'string' && message.length > 0 && message[0].message) {
                extractMessage(message[0].message, args[0], function (result) {
                    backupI(functionName,args,result)
                    callback(result);
                });
            } else {
                logger.info("(" + org + ") " + functionName + ": " + message.toString());
                extractMessage(message, args[0], function (result) {
                    backupI(functionName,args,result)
                    callback(result);
                });
            }
        });
}

var querying = function(args,functionName,callback) {
    query.queryChaincode(getAllPeers(), config.channelName, config.chaincodeName, args, functionName, args[0], org)
        .then(function (message) {
            logger.info("(" + org + ") " + functionName + ": " + message.toString())
            extractMessage(message, args[0], function (result) {
                backupQ(functionName, args,result);
                callback(result);
            });
        });
}

var extractMessage = function(message, usr, callback){
    var response = extractMessageFromText(message);
    if (response && typeof response !== 'string') {
        callback(response);
    } else
    query.getTransactionByID(getAllPeers(), message, usr, config.myOrganisation)
        .then(function(tx) {
            if (tx && typeof tx !== "string") {
                callback({"status": true, "description": "Transaction successful", "trxnId": message});
            } else {
                callback(crErr(message));
            }
        }).catch(function(e) {
            callback(crErr(message));
        })
}

var extractMessageFromText = function(message) {
    var i = -1;
    var i1 = message.indexOf("{");
    var i2 = message.indexOf("[");
    if (i1 >= 0) i = i1;
    if (i2 >= 0 && i2 < i1) i = i2;
    var j = -1;
    var j1 = message.lastIndexOf("}");
    var j2 = message.lastIndexOf("]");
    if (j1 >= 0) j = j1;
    if (j2 >= 0 && j2 > j1) j = j2;
    if(i>=0 && j>=0){
        try {
            var obj = message.substring(i, j + 1);
            var res = JSON.parse(obj);
            return res;
        } catch(e){
            return message;
        }
    } else return message;
}

var getFabricUser = function(login) {
    return helper.getRegisteredUsers(login, org, true);
}


var getAllPeers = function(){
    return Object.keys(helper.ORGS[org].peers);
}


var getAllOrganisations = function(){
    var orgs = [];
    var keys = Object.keys(helper.ORGS);
    if (keys && keys.length>0){
        for(var i=0; i< keys.length; i++){
            if(keys[i]!= "orderer"){
                orgs.push(keys[i]);
            }
        }
    }
    return orgs;
}

var getTrans = function (trxnId, usr, callback) {
    query.getTransactionByID(getAllPeers(), trxnId, usr, config.myOrganisation)
		.then(function(message) {
            callback(message);
		});
}

var getBlockByHash = function (hash, usr) {
	query.getBlockByHash(getAllPeers(), hash, usr, config.myOrganisation).then(
		function(message) {
            return message;
		});
}

var backup = function(func,data,res) {
    var obj = {};
    obj.date = dateFormat(Date.now(), 'yyyy-mm-dd HH:MM:ss.sss');
    obj.func = func;
    obj.data = data;
    obj.result = res;
    return JSON.stringify(obj)+"\n";
}

var backupI = function(func,data,res) {
    fs.appendFile(PathOfBackup.invoke, backup(func,data,res));
}

var backupQ = function(func,data,res) {
    fs.appendFile(PathOfBackup.query, backup(func,data,res));
}

var tryVsData = function (val, val2) {
    if(val && val.length > 0){
        return val;
    } else {
        if(val2) {
            return val2;
        } else return "";
    }
}

module.exports.createUserInBc = createUserInBc;
module.exports.activateUserInBc = activateUserInBc;
module.exports.changeUserPasswordInBc = changeUserPasswordInBc;
module.exports.getUser = getUser;
module.exports.userLogin = userLogin;
module.exports.putOther = putOther;
module.exports.getOther = getOther;
module.exports.createOrderInBc = createOrderInBc;
module.exports.confirmOrder = confirmOrder;
module.exports.repaymentLoan = repaymentLoan;
module.exports.getSumToRepayment = getSumToRepayment;
module.exports.getBorrows = getBorrows;
module.exports.getLends = getLends;
module.exports.getOrder = getOrder;
module.exports.activateOrderInBc = activateOrderInBc;
module.exports.setEthToUserInBc = setEthToUserInBc;
module.exports.tokenTransferInBc = tokenTransferInBc;
module.exports.getHistoryTransferToken = getHistoryTransferToken;
module.exports.getOperations = getOperations;
module.exports.getTrans = getTrans;
module.exports.closeOrder = closeOrder;
module.exports.cancelOrder = cancelOrder;

module.exports.activateOffer = activateOffer;
module.exports.getBorrowOffers = getBorrowOffers;
module.exports.getLendOffers = getLendOffers;
module.exports.createSchedule = createSchedule;
module.exports.setActiveSchedule = setActiveSchedule;
module.exports.getActiveSchedule = getActiveSchedule;
module.exports.getListOfTranchesByDate = getListOfTranchesByDate;
module.exports.getScheduleList = getScheduleList;
module.exports.getScheduleById = getScheduleById;
module.exports.updateOrder = updateOrder;
module.exports.paymentTranche = paymentTranche;
module.exports.getUserTranches = getUserTranches;


setTimeout(function() {
    getUser(config.admins[0].username, function (gu) {
        if(gu && gu.status == false) {
            createUserInBc(config.admins[0].username, config.admins[0].secret, "", "", "", "01.01.1900", "", "", "",
                "ADMIN", "{}", "", "0", "101251", "", "", "", "", "1","", function (usr) {
                    if (usr && usr.status) {
                        activateUs();
                    } else logger.error("ROBOT start error!")
                })
        } else {
            if (gu && gu.status && gu.status != "ACTIVE") {
                activateUs();
            } else {
                logger.info("ROBOT successfully started!")
                robot.startRobot("admin", config.robot_timeout);
            }
        }
    })
}, 120000);


var activateUs = function() {
    activateUserInBc(config.admins[0].username, config.admins[0].secret, function (act) {
        if (act && act.status) {
            logger.info("ROBOT successfully started!")
            robot.startRobot("admin", config.robot_timeout);
        } else logger.error("ROBOT start error!")
    })
}
