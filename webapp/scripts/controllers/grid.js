'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:GridCtrl
 * @description
 * # GridCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('GridCtrl', function ($scope) {
    $scope.awesomeThings = [
      'HTML5 Boilerplate',
      'AngularJS',
      'Karma'
    ];
  });
