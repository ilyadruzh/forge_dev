/**
 * Created by local on 11.11.17.
 */
var api = require('./api.js');
var log4js = require('log4js');
var logger = log4js.getLogger('api');
var dateFormat = require('dateformat');
var config = require('../config.json');


var startRobot = function(user, time){
    logger.info("---------------- START ROBOT ---------------------------------")

    var dateSt = dateFormat(Date.now()-60*60*24*1000*config.robot_count_days, 'yyyy-mm-dd');
    var dateFin = dateFormat(Date.now(), 'yyyy-mm-dd');
    api.getListOfTranchesByDate(user, dateSt, dateFin, function(result){
        if(result && result.result) {
            if(result.result.length>0){
                result.result.forEach(function(tranche){
                    api.getScheduleById(user, tranche.scheduleId, function(schedule){
                        api.getOrder(user, schedule.orderId+"", function(order){
                            api.getUser(order.borrowerId, function(user){
                                var trStart = (new Date(tranche.issueDate)).getTime()+"";
                                var trFinish = (new Date(tranche.repaymentDate)).getTime()+"";
                                api.getHistoryTransferToken(user.login, user.wallet, trStart, trFinish, function(listOfPayments){




                                    logger.info(listOfPayments)// list of transfers

                                    logger.info("---------------- FINISH ROBOT ---------------------------------")
                                })
                            })
                        })
                    })
                })
            } else {
                logger.info(" No data for close tranches")
                logger.info("---------------- FINISH ROBOT ---------------------------------")
            }

        } else logger.error(result);
    })

    setTimeout(function(){
        startRobot(user, time);
    }, time*1000);
}

module.exports.startRobot = startRobot;
