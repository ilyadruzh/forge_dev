var log4js = require('log4js');
var logger = log4js.getLogger('Loan');

var channels = require('./create-channel.js');
var join_ch = require('./join-channel.js');
var install = require('./install-chaincode.js');
var instantiate = require('./instantiate-chaincode.js');
var query = require('./query.js');
var path = require('path');
var config = require('../config.json');


var getInstalledChaincode = function(type){
    query.getInstalledChaincodes("peer0", type, config.admins[0].username, "org1")
        .then(function(message) {
            logger.info("("+type+")peer0, org1: "+message);
        });
    query.getInstalledChaincodes("peer1", type, config.admins[0].username, "org1")
        .then(function(message) {
            logger.info("("+type+")peer1, org1: "+message);
        });
    query.getInstalledChaincodes("peer0", type, config.admins[0].username, "org2")
        .then(function(message) {
            logger.info("("+type+")peer0, org2: "+message);
        });
    query.getInstalledChaincodes("peer1", type, config.admins[0].username, "org2")
        .then(function(message) {
            logger.info("("+type+")peer1, org2: "+message);
        });
}

var installChaincode = function (version) {
    install.installChaincode(["peer0", "peer1"], config.chaincodeName, config.chaincodeName, version, config.admins[0].username, "org1")
        .then(function(message) {
            logger.info("(org1) installChaincode: "+message);
        });
    install.installChaincode(["peer0", "peer1"], config.chaincodeName, config.chaincodeName, version, config.admins[0].username, "org2")
        .then(function(message) {
            logger.info("(org2) installChaincode: "+message);
        });
}

var instantiateChaincode = function (version, org) {
    instantiate.instantiateChaincode(config.channelName, config.chaincodeName, version, config.chaincodeName, '{"Args":["init",""]}', config.admins[0].username, org)
        .then(function (message) {
            logger.info("("+org+") instantiate: " + message);
        });
}

var joinChannel = function () {
    join_ch.joinChannel(config.channelName, ["peer0", "peer1"], config.admins[0].username, "org1")
        .then(function(message) {
            logger.info("(org2) joinChannel: "+message)
        }
    );
    join_ch.joinChannel(config.channelName, ["peer0", "peer1"], config.admins[0].username, "org2")
        .then(function(message) {
                logger.info("(org2) joinChannel: "+message)
            }
        );
}

var createChannel = function() {
    return channels.createChannel(config.channelName, path.join(config.CC_SRC_PATH, "channel/"+config.channelName+".tx"), config.admins[0].username, "org1")
}

var getChannelInfo = function (org) {
    query.getChannels(["peer0", "peer1"], config.admins[0].username, org)
        .then(function (message) {
                logger.info("("+org+") getChannels: " + message)
            }
        );
}

var getChainInfo = function(org){
    query.getChainInfo("peer0", config.admins[0].username, org).then(
        function(message) {
            logger.info("("+org+") getChainInfo: "+message)
        }
    );
}


var configureNetwork = function() {
    createChannel().then(function (channel) {
        joinChannel().then(function (join) {
            installChaincode(config.version).then(function (install) {
                instantiateChaincode(config.version, config.myOrganisation).then(function (instantiate) {

                    logger.info("===============================================================================")
                    logger.info("=================================CHANNEL info =================================")
                    getChannelInfo(config.myOrganisation).then(function() {
                       logger.info("===============================================================================")
                        getChainInfo(config.myOrganisation).then(function () {
                            logger.info("=================================== RESULTS =====================================")
                            logger.info("createChannel: " + channel)
                            logger.info("joinChannel: " + join)
                            logger.info("installChaincode: " + install)
                            logger.info("instantiateChaincode: " + instantiate)
                            logger.info("=============================== Installed CHAINCODE =============================")
                            getInstalledChaincode().then(function () {
                                logger.info("===============================================================================")
                            })
                        })
                    })
                })
            })
        })
    })
}

// configureNetwork();