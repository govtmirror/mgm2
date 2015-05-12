'use strict';

function bin2String(array) {
  var result = "";
  for (var i = 0; i < array.length; i++) {
    result += String.fromCharCode(parseInt(array[i], 2));
  }
  return result;
}

/**
 * @ngdoc service
 * @name mgmApp.mgm
 * @description
 * # mgm
 * Service in the mgmApp.
 */
angular.module('mgmApp').service('mgm', function ($location, $rootScope, $q, $http) {
  console.log("mgm service instantiated");

  var remoteURL = "ws://" + $location.host() + ":" + $location.port() + "/ws";

  self = this;

  self.regions = {};
  self.estates = {};
  self.activeUsers = {};
  self.suspendedUsers = {};
  self.pendingUsers = {};
  self.groups = {};
  self.hosts = {};
  self.jobs = {};
  self.serverConnected = false;

  $rootScope.$on("AuthChange", function (event, value) {
    if (value === false) {
      //we logged out
      self.regions = {};
      self.estates = {};
      self.activeUsers = {};
      self.suspendedUsers = {};
      self.pendingUsers = {};
      self.groups = {};
      self.hosts = {};
      self.serverConnected = false;
    }
  });

  self.connect = function () {
    $rootScope.$broadcast("SyncBegin");
    console.log("Connecting to: " + remoteURL);
    self.ws = new ReconnectingWebSocket(remoteURL);

    self.ws.onopen = function () {
      console.log("Socket has been opened!");
      $rootScope.$broadcast("ServerConnected");
      self.serverConnected = true;
      self.request("GetState", "", function () {
        $rootScope.$broadcast("SyncComplete");
      });
    };

    self.ws.onmessage = function (evt) {
      var message = $.parseJSON(evt.data);
      switch (message.MessageType) {
      case "UserUpdate":
        var user = message.Message;
        if (user.Suspended) {
          self.suspendedUsers[user.UserID] = user;
          if (user.UserID in self.activeUsers) {
            delete self.activeUsers[user.UserID];
          }
        } else {
          self.activeUsers[user.UserID] = user;
          if (user.UserID in self.suspendedUsers) {
            delete self.suspendedUsers[user.UserID];
          }
        }
        $rootScope.$broadcast("UserUpdate", user);
        break
      case "PendingUserUpdate":
        self.pendingUsers[message.Message.UserID] = message.Message;
        $rootScope.$broadcast("PendingUserUpdate", message.Message);
        break;
      case "RegionUpdate":
        message.Message.Status = JSON.parse(message.Message.Status);
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
      case "JobUpdate":
        message.Message.Data = JSON.parse(message.Message.Data)
        self.jobs[message.Message.ID] = message.Message;
        $rootScope.$broadcast("JobUpdate", message.Message);
        break;
      case "ConfigUpdate":
        $rootScope.$broadcast("ConfigUpdate", message.Message);
        break;
      case "HostUpdate":
        self.hosts[message.Message.ID] = message.Message;
        $rootScope.$broadcast("HostUpdate", message.Message);
        break;
      case "HostStatus":
        if( message.Message.ID in self.hosts){
          self.hosts[message.Message.ID].Status = message.Message;
          $rootScope.$broadcast("HostStatusUpdate", message.Message);
        }
        break;
      case "Success":
        var msgID = message.MessageID;
        if (msgID in requestMap) {
          requestMap[msgID].Callback(true, message.Message);
          delete requestMap[msgID];
        } else {
          console.log("Invalid success for nonexistant request: " + msgID);
        }
        break;
      case "Error":
        var msgID = message.MessageID;
        if (msgID in requestMap) {
          requestMap[msgID].Callback(false, message.Message);
          delete requestMap[msgID];
        } else {
          console.log("Invalid error for nonexistant request: " + msgID);
        }
        break;
      default:
        console.log("Error parsing message:");
        console.log(message);
      };

    }

    self.ws.onclose = function (message) {
      console.log("Connection closed");
      self.serverConnected = false;
    }
  };

  self.disconnect = function () {
    self.ws.close();
  };

  /* Request tracking */
  var requestNum = 0;
  var requestMap = {};
  self.request = function (requestType, reqObject, callback) {
    var msgId = requestNum;
    requestNum++;
    requestMap[msgId] = {
      "MessageID": msgId,
      "Callback": callback
    };
    self.ws.send(JSON.stringify({
      "MessageID": msgId,
      "MessageType": requestType,
      "Message": reqObject
    }));
  }

  /* location tracking */
  var locationStack = new Array();
  self.pushLocation = function (url) {
    locationStack.push(url);
  }
  self.popLocation = function () {
    return locationStack.pop();
  }

  self.upload = function (url, file) {
    return $q(function (resolve, reject) {
      var form = new FormData();
      form.append("file", file);
      $http.post(url, form, {
          transformRequest: angular.identity,
          headers: {
            'Content-Type': undefined
          }
        })
        .success(function (data, status, headers, config) {
          resolve();
        })
        .error(function (data, status, headers, config) {
          reject(status);
        });
    });
  }

  /* utility functions */
  self.deleteJob = function(job){
    return $q(function (resolve, reject) {
      self.request("DeleteJob",{
        ID: job.ID
      }, function(success, message){
        if(success){
          delete self.jobs[job.ID];
          resolve();
        } else {
          reject(message);
        }
      })

    });
  }
});
