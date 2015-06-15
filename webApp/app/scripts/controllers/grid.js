'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:GridCtrl
 * @description
 * # GridCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('GridCtrl', function ($scope, mgm, $location, $routeParams, $timeout) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url('/loading');
    }

    $scope.section = $routeParams.section;

    if (!$scope.section) {
      $scope.section = 'estates';
    }

    $scope.isActive = function (section) {
      return this.section === section;
    };

    $scope.estates = mgm.estates;
    $scope.groups = mgm.groups;
    $scope.hosts = mgm.hosts;

    $scope.$on('EstateUpdate', function (event, estate) {
      $timeout(function(){
        if (!(estate.ID in $scope.estates)) {
          $scope.estates[estate.ID] = estate;
        } else {
          if( estate !== $scope.estates[estate.ID]){
            angular.copy(estate, $scope.estates[estate.ID]);
          }
        }
      });
    });

    $scope.$on('GroupUpdate', function (event, group) {
      $timeout(function(){
        if (!(group.ID in $scope.groups)) {
          $scope.groups[group.ID] = group;
        } else {
          if(group !== $scope.groups[group.ID]){
            angular.copy(group, $scope.groups[group.ID]);
          }

        }
      });
    });

    $scope.$on('HostUpdate', function (event, host) {
      $timeout(function(){
        if (!(host.ID in $scope.hosts)) {
          $scope.hosts[host.ID] = host;
        } else {
          if(host !== $scope.hosts[host.ID]){
            angular.copy(host, $scope.hosts[host.ID]);
          }
        }
      });
    });

    $scope.$on('HostStatusUpdate', function (event, stat) {
      if( stat.ID in $scope.hosts){
        $timeout(function(){
          $scope.hosts[stat.ID].Status = stat;
        });
      }
    });

    $scope.getUserNameFromID = function (uuid) {
      if (uuid in mgm.activeUsers) {
        return mgm.activeUsers[uuid].Name;
      }
      if (uuid in mgm.pendingUsers) {
        return mgm.pendingUsers[uuid].Name;
      }
      return '';
    };
    $scope.getUserNamesFromIDs = function (uuids) {
      var users = [];
      for (var i = 0; i < uuids.length; i++) {
        var uuid = uuids[i];
        var name = $scope.getUserNameFromID(uuid);
        if (name !== '') {
          users.push(name);
        }
      }
      return users;
    };
  });
