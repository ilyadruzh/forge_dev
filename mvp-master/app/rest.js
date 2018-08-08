var log4js = require('log4js');
var logger = log4js.getLogger('rest');

var application = require('../app')
var app = application.app;
var helper = require('./helper.js');
var sendMail = require('./sendMail.js');
var api = require('./api.js');
var config = require('../config.json');
var jwt = require('jsonwebtoken');
var hfc = require('fabric-client');
var util = require('util');
const context = "/api";


const DEV = process.env.DEV;
var devToken;

var pathExcluded = [context+'/login',context+'/registerUser', context+'/activateUser'];


app.use(function(req, res, next) {
    var url = req.originalUrl.split("?");
    if ( pathExcluded.indexOf(url[0]) >= 0) {
        return next();
    }
    var token;
    if(DEV=="true"){
        token = devToken;
    } else {
        var tmp = req.get('Authorization');
        if(tmp.indexOf("Bearer") > -1) {
            var val = tmp.split("Bearer ");
            token = val[val.length-1];
        } else token = tmp;
    }

    jwt.verify(token, app.get('secret'), function(err, decoded) {
        if (err) {
            res.send({
                success: false,
                message: 'Failed to authenticate. You need login first.'
            });
            return;
        } else {
            req.username = decoded.username;
            req.orgname = decoded.orgName;
            logger.debug(util.format('Decoded token: username - %s, orgname - %s', decoded.username, decoded.orgName));
            return next();
        }
    });
});

///////////////////////////////////////////////////////////////////////////////
///////////////////////// REST ENDPOINTS START ////////////////////////////////
///////////////////////////////////////////////////////////////////////////////


app.post(context+'/login', function(req, res) {
    var username = req.body.username;
    var password = req.body.password;

    logger.debug('End point : /users');
    logger.debug('User name : ' + username);
    if (!username) {
        res.json(api.getErrorMessage('username'));
        return;
    }
    if (!password) {
        res.json(api.getErrorMessage('password'));
        return;
    }
    api.userLogin(username, password, function (result) {
        if(result && result.status){
            var token = jwt.sign({
                exp: Math.floor(Date.now() / 1000) + parseInt(hfc.getConfigSetting('jwt_expiretime')),
                username: username,
                orgName: config.myOrganisation
            }, app.get('secret'));
            helper.getRegisteredUsers(username, config.myOrganisation, true).then(function(response) {
                if (response && typeof response !== 'string') {
                    response.token = token;
                    result.token = token;
                    if(DEV == "true"){
                        devToken = token;
                    }
                    res.json(result);
                } else {
                    res.json({
                        "status": false,
                        "description": response
                    });
                }
            });
        } else res.json(result);
    })
});

app.put(context+'/changeUserPassword', function(req, res) {
    var username = req.username;
    if(!req.body.oldPassword){
        return res.json(crErr("Old password can't be null"))
    }
    if(!req.body.newPassword){
        return res.json(crErr("New password can't be null"))
    }
    var pass1 = req.body.oldPassword;
    var pass2 = req.body.newPassword;

    if(pass1 == pass2){
        res.json({
            "status": false,
            "description": "oldPassword can't equal with newPassword, try another newPassword"
        });
    } else {
        api.changeUserPasswordInBc(username, pass1, pass2, function(result){
            res.json(result);
        })
    }
});

app.put(context+'/registerUser', function(req, res) {

    if(!req.body.login){
        return res.json(crErr("Login can't be null"))
    }
    if(!req.body.password){
        return res.json(crErr("Password can't be null"))
    }
    if(!req.body.balance){
        return res.json(crErr("Balance can't be null"))
    }
    if(!req.body.role){
        return res.json(crErr("Role can't be null"))
    }
    if(!req.body.email){
        return res.json(crErr("E-mail can't be null"))
    }
    var login                     = req.body.login;
    var pass                      = req.body.password;
    var patronymic                = tryNull(req.body.patronymic);
    var firstName                 = tryNull(req.body.firstName);
    var lastName                  = tryNull(req.body.lastName);
    var birthDate                 = tryNull(req.body.birthDate);
    var passport                  = tryNull(req.body.passport);
    var email                     = tryNull(req.body.email);
    var mobilePhone               = tryNull(req.body.mobilePhone);
    var role                      = tryNull(req.body.role);
    var walletEth                 = tryNull(req.body.wallet);
    var startBalance              = req.body.balance;
    var props                     = tryNull(req.body.props);
    var gender                    = tryNull(req.body.sexId);
    var inn                       = tryNull(req.body.inn);
    var bureauConsentDate         = tryNull(req.body.bureauConsentDate);
    var registrationAddressData   = tryNull(req.body.registrationAddressData);
    var addressData               = tryNull(req.body.addressData);
    var managerId                 = tryNull(req.body.managerId);
    var eth                       = tryNull(req.body.eth);

    if(role == "ADMIN"){
        return res.json({"status":false,"message":"You can't create user with role 'ADMIN'"})
    }

    var token = jwt.sign({
        exp: Math.floor(Date.now() / 1000) + 30*24*60*60,
        username: login,
        password: pass,
        orgName: config.myOrganisation
    }, app.get('secret'));

    var html = "<div>Пожалуйста, не отвечайте на это письмо, оно отправлено автоматически</div>"+
        "<div>Для активации пользователя "+login+" в проекте "+config.chaincodeName+" необходимо перейти по ссылке:</div>"+
        "<div></div>"+
        "<div>"+config.publicHost+context+"/activateUser?value="+token+"</div>"+
        "<div>ссылка будет действительна в течении 30 дней.</div>"+
        "<div>Если Вы не регистрировались в системе, просто удалите это письмо...</div>";


        api.createUserInBc(login,pass,firstName,lastName,patronymic,birthDate,passport,email,mobilePhone,
            role,props,walletEth,startBalance,gender,inn,bureauConsentDate,addressData,
            registrationAddressData,managerId,eth,
            function (result) {

            helper.getRegisteredUsers(login, config.myOrganisation, true).then(function (response) {
                if (response && typeof response !== 'string' && result.status) {

                    sendMail.sendMail(email, config.chaincodeName+' registration USER', html, function(err) {
                        logger.debug("Send e-mail: " + html);
                        if(err) {
                            res.json({
                                "status": false,
                                "description": "You e-mail is not valid or cannot connect to SMTP server!",
                                "message": err
                            });
                        } else res.json(result)
                    })

                } else {
                    if(!result.status){
                        res.json(result)
                    } else
                    res.json({
                        "status": false,
                        "description": response
                    });
                }
            });
        })

})

app.get(context+'/activateUser', function(req, res) {
    var token = req.query.value;
    var username;
    var password;

    jwt.verify(token, app.get('secret'), function(err, decoded) {
        if (err) {
            res.send({
                "status": false,
                "description": "Failed to activate user"
            });
        } else {
            api.activateUserInBc(decoded.username, decoded.password, function(result){
                res.json(result);
            })
        }
    });
})


app.put(context+'/tokenTransfer', function(req, res) {
    var username = req.username;

    var tx     = tryNull(req.body.tx);
    var from   = tryNull(req.body.from);
    var to     = tryNull(req.body.to);
    var amount = tryNull(req.body.amount);
    var orderId = tryNull(req.body.orderId);

    if(from && to && amount){
        api.tokenTransferInBc(username, tx, from, to, amount,orderId, function(result){
            res.json(result);
        })
    } else {
        res.json({
            "status": false,
            "description": "cannot set necessary parameters"
        });
    }
});


app.get(context+'/getTransfers', function(req, res) {
    var username = req.username;

    var eth    = tryNull(req.query.eth);
    var start  = tryNull(req.query.start);
    var finish = tryNull(req.query.finish);

    api.getHistoryTransferToken(username, eth, start, finish, function(result){
        res.json(result);
    })
});

app.get(context+'/getOperations', function(req, res) {
    var username = req.username;

    var contractName = tryNull(req.query.filter);
    var start        = tryNull(req.query.start);
    var finish       = tryNull(req.query.finish);

    api.getOperations(username, contractName, start, finish, function(result){
        res.json(result);
    })
});

app.get(context+'/getUser', function(req, res) {
    var username = req.username;

    api.getUser(username, function(result){
        res.json(result);
    })
});

app.put(context+'/setUserEth', function(req, res) {
    var username = req.username;

    var eth = tryNull(req.body.eth);
    var wallet = tryNull(req.body.wallet);
    api.setEthToUserInBc(username,wallet,eth, function(result){
        res.json(result);
    })
});

//-------------- Order ---------------------------------------
app.put(context+'/createOrder', function(req, res) {
    var username = req.username;

    var period       = tryNull(req.body.period);
    // var borrower     = tryNull(req.body.borrower);
    // var lender       = tryNull(req.body.lender);
    var sum          = tryNull(req.body.sum);
    // var percent      = tryNull(req.body.percent);
    var startDate    = tryNull(req.body.startDate);
    var finishDate   = tryNull(req.body.finishDate);


    // Идем в брейнсофт за % ставкой

    api.createOrderInBc(username,period,username,"",sum,"-1",startDate,finishDate,function(result){
        res.json(result);
    })
});

app.post(context+'/getOrder', function(req, res) {
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.getOrder(username,orderId,function(result){
        res.json(result);
    })
});

app.put(context+'/confirmOrder', function(req, res) {
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.confirmOrder(username,orderId,function(result){
        res.json(result);
    })
});

app.put(context+'/paymentTranche', function(req, res) {
    var username = req.username;
    var scheduleId = tryNull(req.body.scheduleId);
    var dateOfTranche = tryNull(req.body.dateOfTranche);// format 2017-11-01
    var eachRepaymentFee = tryNull(req.body.eachRepaymentFee);// float64

    api.paymentTranche(username,dateOfTranche,scheduleId,eachRepaymentFee,function(result){
        res.json(result);
    })
});

app.put(context+'/activateOrder', function(req, res) {
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.activateOrderInBc(username,orderId,function(result){
        res.json(result);
    })
});

app.put(context+'/activateOffer', function(req, res) {
    var username = req.username;
    var orderId = tryNull(req.body.orderId);


    api.activateOffer(username, orderId, function(result){
        res.json(result);
    })
});

app.post(context+'/getBorrowOffers', function(req, res) {
    var username = req.username;
    var sumSt = tryNull(req.body.sumSt);
    var percentSt = tryNull(req.body.percentSt);
    var periodSt = tryNull(req.body.periodSt);
    var sumFin = tryNull(req.body.sumFin);
    var percentFin = tryNull(req.body.percentFin);
    var periodFin = tryNull(req.body.periodFin);
    api.getBorrowOffers(username, sumSt, percentSt, periodSt, sumFin, percentFin, periodFin, function(result){
        res.json(getMinMax(result));
    })
});

app.post(context+'/closeOrder', function(req, res) {
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.closeOrder(username, orderId, function(result){
        res.json(result);
    })
});

app.post(context+'/cancelOrder', function(req, res) {
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.cancelOrder(username, orderId, function(result){
        res.json(result);
    })
});

app.post(context+'/getBorrows', function(req, res) {
    var username = req.username;
    api.getBorrows(username, username, function(result){
        res.json(result);
    })
});

app.post(context+'/getLends', function(req, res) {
    var username = req.username;
    api.getLends(username, username, function(result){
        res.json(result);
    })
});


app.post(context+'/getTransaction', function(req, res) {
    var username = req.username;
    var trans = tryNull(req.body.trans);

    api.getTrans(trans, username, function(result){
        res.json(result);
    })
});

app.post(context+'/getLendOffers', function(req, res) {
    var username = req.username;

    var sumSt = tryNull(req.body.sumSt);
    var percentSt = tryNull(req.body.percentSt);
    var periodSt = tryNull(req.body.periodSt);
    var sumFin = tryNull(req.body.sumFin);
    var percentFin = tryNull(req.body.percentFin);
    var periodFin = tryNull(req.body.periodFin);

    api.getLendOffers(username, sumSt, percentSt, periodSt, sumFin, percentFin, periodFin, function(result){
        res.json(getMinMax(result));
    })
});

app.put(context+'/updateOrder', function(req, res) {
    var username = req.username;

    var period       = tryNull(req.body.period);
    var borrower     = tryNull(req.body.borrower);
    var lender       = tryNull(req.body.lender);
    var sum          = tryNull(req.body.sum);
    var percent      = tryNull(req.body.percent);
    var startDate    = tryNull(req.body.startDate);
    var finishDate   = tryNull(req.body.finishDate);
    var orderId      = tryNull(req.body.orderId);
    var contractId   = tryNull(req.body.contractId);
    if (contractId == ""){
        contractId = "0";
    }

    api.updateOrder(username,period,borrower,lender,sum,percent,startDate,finishDate,orderId,contractId,function(result){
        res.json(result);
    })
});

//-------------- Schedule ---------------------------------------
app.put(context+'/createSchedule', function(req, res) {
    var username = req.username;

    var orderId        = tryNull(req.body.orderId);
    if(!orderId){
        res.json(crErr("Can't be null orderId"))
    }
    api.createSchedule(username,orderId,function(result){
        res.json(result);
    })
});

app.put(context+'/setActiveSchedule', function(req, res) {
    // 0 - Login,
    // 1 - ScheduleId,
    // 2 - OrderId
    var username = req.username;
    var orderId = tryNull(req.body.orderId);
    var scheduleId = tryNull(req.body.scheduleId);

    api.setActiveSchedule(username, scheduleId, orderId, function (result) {
        res.json(result);
    })
});

app.post(context+'/getActiveSchedule', function(req, res) {
    // 0 - Login, 1 - OrderId
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.getActiveSchedule(username, orderId, function (result) {
        res.json(removeScId(result));
    })
});

app.post(context+'/getListOfTranchesByDate', function(req, res) {
    // 0 - Login,
    // 1 - dateStart  format '2017-06-02'
    // 2 - dateFinish format '2017-07-02'
    var username = req.username;
    var dateStart = tryNull(req.body.dateStart);
    var dateFinish = tryNull(req.body.dateFinish);

    api.getListOfTranchesByDate(username, dateStart, dateFinish, function (result) {
        if(result && result.result) {
            if(result.result.length > 0) {
                for(var j=0;j<result.result.length;j++) {
                    var resSc = result.result[j];
                    delete resSc.scheduleId;
                    result.result[j] = resSc;
                }
            }
        }
        res.json(result);
    })
});

app.post(context+'/getScheduleList', function(req, res) {
    // 0 - Login, 1 - OrderId
    var username = req.username;
    var orderId = tryNull(req.body.orderId);

    api.getScheduleList(username, orderId, function (result) {
        if(result) {
            if(result.length > 0) {
                for(var j=0;j<result.length;j++) {
                    result[j] = removeScId(result[j]);
                }
            }
        }
        res.json(result);
    })
});

app.post(context+'/getScheduleById', function(req, res) {
    // 0 - Login, 1 - scheduleId
    var username = req.username;
    var scheduleId = tryNull(req.body.scheduleId);

    api.getScheduleById(username, scheduleId, function (result) {
        res.json(removeScId(result));
    })
});

app.post(context+'/getUserTranches', function(req, res) {
    // 0 - Login, 1 - scheduleId
    var username = req.username;
    var dateStart = tryNull(req.body.dateStart);
    var dateFinish = tryNull(req.body.dateFinish);

    api.getUserTranches(username, dateStart, dateFinish, function (result) {
        res.json(removeScId(result));
    })
});

///////////////////////////////////////////////////////////////////////////////
app.get(context+'/getOther', function(req, res) {
    var username = req.username;

    var key = req.query.key;

    api.getOther(username,key,function(result){
        res.json(result);
    })
});

removeScId = function(result){
    if(result) {
        if (result.tranches) {
            if (result.tranches.length > 0) {
                for (var i = 0; i < result.tranches.length; i++) {
                    delete result.tranches[i].scheduleId;
                }
            }
        }
    }
    return result;
}

tryNull = function(val){
    if(!val){
        return "";
    } else return val;
}

crErr = function(text) {
    return {"status":false,"description":text};
}

getMinMax = function(result){
    var sMin = 0, sMax = 0;
    var perMin = 0, perMax = 0;
    var prMin = 0, prMax = 0;
    if (result && result.result && result.result.length > 0){
        for (var i=0;i<result.result.length;i++){
            if(sMin == 0) sMin = result.result[i].sum;
            if(prMin == 0) prMin = result.result[i].percent;
            if(perMin == 0) perMin = result.result[i].period;
            sMin = Math.min(sMin, result.result[i].sum)
            perMin = Math.min(perMin, result.result[i].period)
            prMin = Math.min(prMin, result.result[i].percent)
            sMax = Math.max(sMax, result.result[i].sum)
            perMax = Math.max(perMax, result.result[i].period)
            prMax = Math.max(prMax, result.result[i].percent)
        }
    }
    result.periodMin = perMin;
    result.periodMax = perMax;
    result.procentMin = prMin;
    result.procentMax = prMax;
    result.sumMin = sMin;
    result.sumMax = sMax;
    return result;
}

require("./unit-test.js")
