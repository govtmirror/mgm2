'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionconsoleCtrl
 * @description
 * # RegionconsoleCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionconsoleCtrl', function ($scope, $modalInstance, $timeout, region) {

    $scope.region = region;
    $scope.lines = []

    $scope.close = function() {
      $modalInstance.close();
    };

    $scope.$on('ConsoleUpdate', function(event, status){
      if( status.UUID == region.UUID){
        $timeout(function(){
          $scope.lines.push(status.Line);
        })
      }
    });

  });
