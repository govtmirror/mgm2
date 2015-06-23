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
    $scope.estates = mgm.estates;
    $scope.hosts = mgm.hosts;
    $scope.hosts[0] = {
      ID: 0,
      Hostname: '<none>'
    };

    $scope.currentHost = '';
    $scope.currentEstate = '';
    $scope.currentX = region.LocX;
    $scope.currentY = region.LocY;
    $scope.errorMsg = '';

    function init(){
      $scope.currentHost = $scope.hosts[$scope.region.Host];
      for (var id in $scope.estates) {
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
      if($scope.currentX === undefined){
        $scope.errorMsg = 'X coordinate is invalid';
        return;
      }
      if($scope.currentY === undefined){
        $scope.errorMsg = 'Y coordinate is invalid';
        return;
      }
      $scope.errorMsg = '';
      if($scope.currentX !== $scope.region.LocX || $scope.currentY !== $scope.region.LocY){
        console.log('Set x,y to: ' + $scope.currentX + ', ' + $scope.currentY);
      }
    };

    $scope.setEstate = function() {
      if ($scope.currentEstate.Name !== region.Estate) {
        console.log('Set estate to: ' + $scope.currentEstate.Name);
      }
    };

    $scope.setHost = function() {
      if ($scope.currentHost.ID !== region.Host) {
        console.log('Set host to: ' + $scope.currentHost.Hostname);
      }
    };

    $scope.$on('RegionUpdate', function(event, r){
      if(region.UUID === r.UUID){
        $timeout(function(){
          $scope.region = r;
          init();
        });
      }
    });

    init();
  });
