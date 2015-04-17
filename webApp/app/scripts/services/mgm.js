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

  self.regions = {}
  self.estates = {}
  self.users = {}
  self.groups = {}
  self.hosts = {}

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
        self.users[message.Message.UserID] = message.Message;
        $rootScope.$broadcast("UserUpdate", message.Message);
        break
      case "RegionUpdate":
        self.regions[message.Message.UUID] = message.Message;
        $rootScope.$broadcast("RegionUpdate", message.Message);
        break;
      case "EstateUpdate":
        self.estates[message.Message.ID] = message.Message;
        $rootScope.$broadcast("EstateUpdate", message.Message);
        break;
      case "GroupUpdate":
        self.groups[message.Message.ID] = message.Message;
        $rootScope.$broadcast("GroupUpdate", message.Message);
        break;
      case "HostUpdate":
        message.Message.Status = JSON.parse(message.Message.Status);
        self.hosts[message.Message.Address] = message.Message;
        $rootScope.$broadcast("HostUpdate", message.Message);
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
