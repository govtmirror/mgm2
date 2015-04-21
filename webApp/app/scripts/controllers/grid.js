'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:GridCtrl
 * @description
 * # GridCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('GridCtrl', function ($scope, mgm, $location, $routeParams) {

    if ($scope.auth === undefined) {
      mgm.pushLocation($location.url());
      $location.url("/login");
    }

    $scope.section = $routeParams["section"];

    if (!$scope.section) {
      $scope.section = "estates";
    }

    $scope.isActive = function (section) {
      return this.section === section;
    }

    $scope.estates = mgm.estates;
    $scope.groups = mgm.groups;
    $scope.hosts = mgm.hosts;

    $scope.$on("EstateUpdate", function (event, estate) {
      if (!(estate.ID in $scope.estates)) {
        $scope.estates[estate.ID] = estate;
      } else {
        angular.copy(estate, $scope.estates[estate.ID]);
      }
    });

    $scope.$on("GroupUpdate", function (event, group) {
      if (!(group.ID in $scope.groups)) {
        $scope.groups[group.ID] = group;
      } else {
        angular.copy(group, $scope.groups[group.ID]);
      }
    });

    $scope.$on("HostUpdate", function (event, host) {
      if (!(host.Address in $scope.hosts)) {
        $scope.hosts[host.Address] = host;
      } else {
        angular.copy(host, $scope.hosts[host.Address]);
      }
    });


    $scope.getUserNameFromID = function (uuid) {
      if (uuid in mgm.activeUsers) {
        return mgm.activeUsers[uuid].Name
      }
      if (uuid in mgm.pendingUsers) {
        return mgm.pendingUsers[uuid].Name
      }
      return ""
    }
    $scope.getUserNamesFromIDs = function (uuids) {
      var users = [];
      for (var i = 0; i < uuids.length; i++) {
        var uuid = uuids[i];
        var name = $scope.getUserNameFromID(uuid);
        if (name !== "") {
          users.push(name);
        }
      }
      return users;
    }

    $scope.TimestampToDate = function (timestamp) {
      if (timestamp == undefined || timestamp == "") {
        return "~";
      }
      var last = new Date(timestamp * 1000);
      var seconds = Math.floor(((new Date()).getTime() - last.getTime()) / 1000);

      var numdays = Math.floor(seconds / 86400);
      if (numdays > 0) {
        return numdays + " days ago";
      }
      var numhours = Math.floor((seconds % 86400) / 3600);
      if (numhours > 0) {
        return numhours + " hours ago";
      }
      var numminutes = Math.floor(((seconds % 86400) % 3600) / 60);
      return numminutes + " minutes ago";
    }
  });
