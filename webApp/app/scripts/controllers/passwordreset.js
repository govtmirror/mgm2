'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:PasswordresetCtrl
 * @description
 * # PasswordresetCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('PasswordresetCtrl', function ($scope, mgmPublic) {
    $scope.stage = '';

    $scope.input = {
      email: '',
      uname: '',
      token: '',
      password: '',
      confirm: ''
    };

    $scope.error = {
      email: '',
      uname: '',
      token: '',
      password: '',
      other: ''
    };

    $scope.resetErrors = function () {
      $scope.error.uname = '';
      $scope.error.email = '';
      $scope.error.token = '';
      $scope.error.password = '';
      $scope.error.other = '';
    };

    $scope.needCode = function () {
      $scope.resetErrors();
      $scope.stage = 'NeedCode';
    };

    $scope.haveCode = function () {
      $scope.resetErrors();
      $scope.stage = 'HaveCode';
    };

    $scope.getCode = function () {
      $scope.resetErrors();

      var email = $scope.input.email.trim();
      if (email === '') {
        $scope.error.email = 'Email is required';
        return;
      }
      if (!/(.+)@(.+){2,}\.(.+){2,}/.test(email)) {
        $scope.error.email = 'Invalid email format';
        return;
      }
      $scope.error.email = '';

      $scope.stage = 'Processing';

      mgmPublic.requestResetCode(email).then(
        function () {
          //success
          $scope.stage = 'RequestComplete';
        },
        function (msg) {
          //failure
          $scope.error.other = msg;
          $scope.stage = 'NeedCode';
        }
      );
    };

    $scope.resetPassword = function () {
      $scope.resetErrors();

      var uname = $scope.input.uname.trim();
      var token = $scope.input.token.trim();
      var pword = $scope.input.password.trim();
      var confirm = $scope.input.confirm.trim();

      // Validate user name
      if (uname === '') {
        $scope.error.uname = 'User name is required';
        return;
      }
      if (uname.split(' ').length !== 2) {
        $scope.error.uname = '"firstname lastname" only';
        return;
      }

      //Validate token format
      if (token === '') {
        $scope.error.token = 'Token is required';
        return;
      }
      //token is a guid
      if (!/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/.test(token)) {
        $scope.error.token = 'Invalid token format';
        return;
      }

      //Validate password and password confirmation
      if (pword === '') {
        $scope.error.password = 'Password is required';
        return;
      }
      if (pword !== confirm) {
        $scope.error.password = 'Passwords do not match';
        return;
      }

      $scope.stage = 'Processing';

      //submit password change to service
      mgmPublic.resetPassword(uname, token, pword).then(
        function () {
          //success
          $scope.stage = 'ResetComplete';
          $scope.input.uname = "";
          $scope.input.token = "";
          $scope.input.password = "";
          $scope.input.confirm = "";
        },
        function (msg) {
          $scope.error.other = msg;
          $scope.stage = 'HaveCode';
        }
      );
    };
  });
