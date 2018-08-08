var RKNContract = artifacts.require("./RKNContract.sol");

module.exports = function(deployer) {
  deployer.deploy(RKNContract);
};
