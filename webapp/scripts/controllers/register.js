'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegisterCtrl
 * @description
 * # RegisterCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegisterCtrl', function ($scope, MgmPublic) {
    $scope.stage = 'Eupa';

    $scope.acceptEupa = function () {
      $scope.stage = 'Account';
    };

    $scope.error = {
      fname: '',
      lname: '',
      email: '',
      pword: '',
      gender: '',
      other: ''
    };

    $scope.registrant = {
      fname: '',
      lname: '',
      email: '',
      pword: '',
      pword2: '',
      gender: '',
      reasons: ''
    };

    $scope.resetErrors = function () {
      $scope.error.fname = '';
      $scope.error.lname = '';
      $scope.error.email = '';
      $scope.error.pword = '';
      $scope.error.gender = '';
      $scope.error.other = '';
    };

    $scope.apply = function () {
      $scope.resetErrors();
      // User input validation
      var fname = $scope.registrant.fname.trim();
      var lname = $scope.registrant.lname.trim();
      var email = $scope.registrant.email.trim();
      var pword = $scope.registrant.pword.trim();
      var gender = $scope.registrant.gender;

      if (fname === '') {
        $scope.error.fname = 'Name is required';
        return;
      }
      if (fname.split(' ').length !== 1) {
        $scope.error.fname = 'First name only';
        return;
      }
      if (lname === '') {
        $scope.error.lname = 'Name is required';
        return;
      }
      if (lname.split(' ').length !== 1) {
        $scope.error.lname = 'Last name only';
        return;
      }
      if (email === '') {
        $scope.error.email = 'Email is required';
        return;
      }
      if (!/(.+)@(.+){2,}\.(.+){2,}/.test(email)) {
        $scope.error.email = 'Invalid email entered';
        return;
      }
      if (pword === '') {
        $scope.error.pword = 'A password is required';
        return;
      }
      if (pword !== $scope.registrant.pword2) {
        $scope.error.pword = 'Passwords do not match';
        return;
      }
      if (gender !== 'M' && gender !== 'F') {
        $scope.error.gender = 'Select your gender';
        return;
      }

      console.log('User input validated');

      $scope.state = 'Processing';
      var uname = fname + ' ' + lname;
      var deferred = MgmPublic.registerUser(uname, email, gender, pword, $scope.registrant.reasons);
      deferred.then(
        function () {
          $scope.state = 'Registered';
        },
        function (msg) {
          
          $scope.error.other = msg;
        }
      );
    };
  });