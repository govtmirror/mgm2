'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:LoginCtrl
 * @description
 * # LoginCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('LoginCtrl', function ($scope, mgmPublic) {
    $scope.stage = 'Standby';

    $scope.user = {
      uname: '',
      password: ''
    };

    $scope.error = {
      uname: '',
      password: '',
      other: ''
    };

    $scope.clearErrors = function () {
      $scope.error.uname = '';
      $scope.error.password = '';
      $scope.error.other = '';
    };

    $scope.login = function () {
      $scope.clearErrors();
      var uname = $scope.user.uname.trim();
      var pword = $scope.user.password.trim();
      if (uname === '') {
        $scope.error.uname = 'User name required';
        return;
      }
      if (pword === '') {
        $scope.error.password = 'Password required';
        return;
      }
      console.log('login function' + $scope.user.uname);
      mgmPublic.login(uname, pword).then(
        function () {
          console.log('login success');
          $scope.user.uname = '';
          $scope.user.password = '';
        },
        function (msg) {
          $scope.error.other = msg;
        }
      );
    };
  });
