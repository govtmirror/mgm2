'use strict';

/**
 * @ngdoc service
 * @name mgmApp.mgm
 * @description
 * # mgm
 * Service in the mgmApp.
 */
angular.module('mgmApp').service('mgm', function ($location, $rootScope) {
  console.log("mgm service instantiated");

  var remoteURL = "ws://" + $location.host() + ":" + $location.port() + "/ws";

  self = this;

  self.account = {
    UserID: "",
    Name: "",
    Email: "",
    AccessLevel: 0
  };

  self.regions = {}

  this.connect = function () {
    console.log("Connecting to: " + remoteURL);
    self.ws = new ReconnectingWebSocket(remoteURL);

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

    self.ws.onmessage = function (evt) {
      var message = $.parseJSON(evt.data);
      switch (message.MessageType) {
      case "AccountUpdate":
        self.account = message.Message
        $rootScope.$broadcast("AccountUpdate");
        break
      case "RegionUpdate":
        self.regions[message.Message.UUID] = message.Message;
        $rootScope.$broadcast("RegionUpdate", message.Message);
        break;
      default:
        console.log("Error parsing message:");
        console.log(message);
      };

    }

    self.ws.onclose = function (message) {
      console.log("Connection closed");
    }
  };

  this.disconnect = function () {
    self.ws.close();
  };

});
