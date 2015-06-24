'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:ManageregionCtrl
 * @description
 * # ManageregionCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('ManageregionCtrl', function($scope, $modalInstance, $timeout, mgm, region) {

    $scope.region = region;
    $scope.estates = [];
    $scope.hosts = mgm.hosts;
    $scope.hosts[0] = {
      ID: 0,
      Hostname: '<none>'
    };

    $scope.currentHost = '';
    $scope.currentEstate = '';
    $scope.currentX = region.LocX;
    $scope.currentY = region.LocY;

    function init() {
      $scope.estates = [];
      for (var e in mgm.estates){
        $scope.estates.push(mgm.estates[e]);
      }

      $scope.currentHost = $scope.hosts[$scope.region.Host];
      for (var id in $scope.estates) {
        var estate = $scope.estates[id];
        for (var i = 0; i < $scope.estates[id].Regions.length; i++) {
          if ($scope.estates[id].Regions[i] === $scope.region.UUID) {
            $scope.currentEstate = $scope.estates[id];
            region.Estate = $scope.estates[id].Name;
          }
        }
      }
    }

    $scope.close = function() {
      $modalInstance.close();
    };

    $scope.setXY = function() {
      if ($scope.currentX === undefined) {
        alertify.error('X coordinate is invalid');
        return;
      }
      if ($scope.currentY === undefined) {
        alertify.error('Y coordinate is invalid');
        return;
      }
      if ($scope.currentX !== $scope.region.LocX || $scope.currentY !== $scope.region.LocY) {
        alertify.log('Set x,y to: ' + $scope.currentX + ', ' + $scope.currentY);
      }
    };

    $scope.setEstate = function() {
      if ($scope.currentEstate.Name !== region.Estate) {
        mgm.request('SetEstate', {
          'RegionUUID': region.UUID,
          'ID': $scope.currentEstate.ID
        }, function(success, msg) {
          alertify.log('' + success + ": " + msg);

        })
      }
    };

    $scope.setHost = function() {
      if ($scope.currentHost.ID !== region.Host) {
        alertify.log('Set host to: ' + $scope.currentHost.Hostname);
      }
    };

    $scope.$on('EstateUpdate', function(event, estate) {
      $timeout(function() {
        init();
      });
    });

    $scope.$on('RegionUpdate', function(event, r) {
      if (region.UUID === r.UUID) {
        $timeout(function() {
          $scope.region = r;
          init();
        });
      }
    });

    init();
  });
