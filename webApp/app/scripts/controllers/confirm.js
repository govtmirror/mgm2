'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:ConfirmCtrl
 * @description
 * # ConfirmCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('ConfirmCtrl', function ($scope, $modalInstance, object) {

    $scope.object = object;

    $scope.confirm = function(){
      $modalInstance.close();
    }

    $scope.cancel = function() {
      $modalInstance.dismiss();
    };
  });
