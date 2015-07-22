'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionconsoleCtrl
 * @description
 * # RegionconsoleCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionconsoleCtrl', function ($scope, $modalInstance, $timeout, mgm, region) {

    $scope.region = region;
    $scope.lines = []

    $scope.close = function() {
      $modalInstance.close();
    };

    $scope.$on('ConsoleInput', function(event, line){
      mgm.request('ConsoleCommand', {
        Message: line
      }, function(success, msg) {
        if (success) {
          //alertify.success(msg);
        } else {
          alertify.error(msg);
        }
      });
    });

    $scope.$on('ConsoleUpdate', function(event, status){
      if( status.UUID == region.UUID){
        $timeout(function(){
          $scope.lines.push(status.Line);
        })
      }
    });

  });
