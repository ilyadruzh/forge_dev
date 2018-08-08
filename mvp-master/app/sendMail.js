var config = require('../config.json');
var nodemailer = require('nodemailer');
var jwt = require('jsonwebtoken');
var transport;

var transport = nodemailer.createTransport({
    host: 'smtp.gmail.com',
    port: 465,
    secure: true, // use SSL
    auth: {
        user: 'sbt.blockchain@gmail.com',
        pass: new Buffer("YXp4YXp4MTIzQA==", 'base64').toString('ascii')
    }
});

// var transport = nodemailer.createTransport({direct:true,
//     host: 'smtp.yandex.ru',
//     port: 465,
//     auth: {
//         user: 'sbt.blockchain@yandex.ru',
//         pass: new Buffer("YXp4YXp4MTIzQA==", 'base64').toString('ascii')
//     },
//     secure: true
// });


var sendMail = function(email, subj, html, callback){

    var mailOptions = {
        from: 'sbt.blockchain@gmail.com',
        to: email,
        subject: subj,
        html: html
    }
    transport.sendMail(mailOptions, function(err, info) {
            if (err) {
                callback(err);
            } else callback(null);
        }
    )};


module.exports.sendMail = sendMail;