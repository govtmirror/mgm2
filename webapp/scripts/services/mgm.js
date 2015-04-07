'use strict';

/**
 * @ngdoc service
 * @name mgmApp.mgm
 * @description
 * # mgm
 * Service in the mgmApp.
 */
angular.module('mgmApp').service('mgm', function ($location) {

  window.WebSocket = window.WebSocket || window.MozWebSocket;
  
  var remoteURL = "ws://"+$location.host() + ":" + $location.port() + "/ws";
  console.log("Connecting to: " + remoteURL);
  var ws = new WebSocket(remoteURL);

  ws.onopen = function () {
    console.log("Socket has been opened!");
  };

});