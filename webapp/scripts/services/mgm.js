'use strict';

/**
 * @ngdoc service
 * @name mgmApp.mgm
 * @description
 * # mgm
 * Service in the mgmApp.
 */
angular.module('mgmApp').service('mgm', function ($location) {
  console.log("mgm service instantiated");

  window.WebSocket = window.WebSocket || window.MozWebSocket;

  var remoteURL = "ws://" + $location.host() + ":" + $location.port() + "/ws";
  console.log("Connecting to: " + remoteURL);
  var ws = new WebSocket(remoteURL);

  ws.onopen = function () {
    console.log("Socket has been opened!");
    var testMessage = {
      "MessageType": "TestMessage",
      "Message": {
        "first": 1,
        "second": 2
      }
    };
    console.log("Sending " + testMessage);
    ws.send(JSON.stringify(testMessage));
  };

  ws.onmessage = function (message) {
    console.log("message received:");
    console.log(message);
  }

  ws.onclose = function (message) {
    console.log("Connection closed");
  }

});