'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:GridCtrl
 * @description
 * # GridCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('GridCtrl', function ($scope, mgm, $location, $routeParams, $timeout, $modal) {

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

    $scope.host = {
      delete: function(host){
        host.RunningRegions = 0;
        if(host.Regions != null){
          var regions = mgm.regions;
          for(var uuid in host.Regions){
            if(regions[uuid].Running){
              host.RunningRegions++;
            }
          }
        }
        var modInst = $modal.open({
          animation: false,
          templateUrl: 'removeHostModal.html',
          backdrop: 'static',
          keyboard: false,
          controller: 'ConfirmCtrl',
          resolve: {
            object: function() {
              return host;
            },
          }
        });
        modInst.result.then(function() {
          console.log('remove host confirmed');
          mgm.request('RemoveHost', {'ID' : host.ID}, function(done, message){
            if(done){
              //host is deleted
              console.log("Success: " + message);
            } else {
              //an error occurred
              console.log("Error: " + message);
            }
          });
        });
      },
      countRunning: function(host){
        var running = 0;
        var regions = mgm.regions;
        for(var uuid in host.Regions) {
          console.log(uuid);
        }
        return ""+ running;
      }

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
