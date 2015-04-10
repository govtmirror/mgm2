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

  self = this;
  
  this.connect = function () {
    console.log("Connecting to: " + remoteURL);
    self.ws = new WebSocket(remoteURL);

    self.ws.onopen = function () {
      console.log("Socket has been opened!");
      var testMessage = {
        "MessageType": "TestMessage",
        "Message": {
          "first": 1,
          "second": 2
        }
      };
      console.log("Sending " + testMessage);
      self.ws.send(JSON.stringify(testMessage));
    };

    self.ws.onmessage = function (message) {
      console.log("message received:");
      console.log(message);
    }

    self.ws.onclose = function (message) {
      console.log("Connection closed");
    }
  };

  this.disconnect = function () {
    self.ws.close();
  };

});