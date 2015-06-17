'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:ManageregionCtrl
 * @description
 * # ManageregionCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('ManageregionCtrl', function ($scope, $modalInstance, region, estates, hosts) {

    $scope.region = region;
    $scope.estates = estates;
    $scope.hosts = hosts;

    $scope.close = function() {
      $modalInstance.close();
    };

    $scope.setXY = function(x, y){
      console.log('Set x,y to: ' + x + ', ' + y);
    };

    $scope.setEstate = function(estate){
      console.log('Set estate to: ' + estate.Name);
    };

    $scope.setHost = function(host){
      console.log('Set host to: ' + host.Hostname);
    };

  });
