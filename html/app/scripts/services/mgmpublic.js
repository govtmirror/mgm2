'use strict';

/**
 * @ngdoc service
 * @name mgmApp.RegisterService
 * @description
 * # RegisterService
 * Service in the mgmApp.
 */
angular.module('mgmApp')
  .service('mgmPublic', function ($q, $http, $rootScope, mgm, $timeout, $localStorage) {

    var self = this;
    var token = null;
    var tokenTimer = null;
    self.loggedIn = false;

    var $store = $localStorage.$default({
      valid: false
    });

    //if we have a saved token, attempt to resume session
    if ($store.valid) {
      $timeout(function(){
        console.log('Resuming connection');
        $rootScope.auth = {
          UUID: $store.uuid,
          AccessLevel: $store.accessLevel
        };
        mgm.token = $store.token
        $rootScope.$broadcast('AuthChange', true);
      });
    }

    function refreshToken(){
      mgm.ws.emit('RefreshToken', '', function(data){
        if(data.Success){
          $store.token = data.Token;
          mgm.token = data.Token;
        }
      })
    }

    // token expires after 60 minutes, refresh every 30
    $rootScope.$on('ServerConnected', function(){
      //shortcut a refresh now, in case we are refreshing often and this timer is not firing
      refreshToken();
      tokenTimer = setInterval(refreshToken, 1800000);
    })

    this.logout = function () {
      self.loggedIn = false;
      $rootScope.auth = {};
      mgm.token = '';
      $store.valid = false;
      delete $store.uuid;
      delete $store.accessLevel;
      delete $store.token;
      clearTimeout(tokenTimer);
      $rootScope.$broadcast('AuthChange', false);
    }

    this.login = function (username, password) {

      return $q(function (resolve, reject) {

        console.log('Authenticating...');
        $http.post('/auth', {
          'username': username,
          'password': password
        }).success(function (data /*,status, headers, config*/) {
          if (data.Success) {
            self.loggedIn = true;
            $rootScope.auth = {
              UUID: data.UUID,
              AccessLevel: data.AccessLevel
            };
            $store.valid = true;
            $store.uuid = data.UUID;
            $store.accessLevel = data.AccessLevel;
            $store.token = data.Token;
            mgm.token = data.Token;
            $rootScope.$broadcast('AuthChange', true);
            resolve('login successfull');
          } else {
            reject(data.Message);
            console.log(data);
          }
        }).error(function (/*data, status, headers, config*/) {
          console.log('an error occurred');
          reject('Error connecting to MGM');
        });
      });
    };

    this.registerUser = function (uname, email, template, password, reason) {
      return $q(function (resolve, reject) {
        $http.post('/auth/register', {
          'Name': uname,
          'Email': email,
          'Template': template,
          'Password': password,
          'Summary': reason,
        }).success(function (data /*, status, headers, config*/) {
          if (data.Success) {
            resolve('Registration Successful');
          } else {
            reject(data.Message);
            console.log(data);
          }
        }).error(function (/*data, status, headers, config*/) {
          console.log('an error occurred');
          reject('Error connecting to MGM');
        });
      });
    };

    this.requestResetCode = function (email) {
      return $q(function (resolve, reject) {
        $http.post('/auth/passwordToken', {
          'Email': email
        }).success(function (data /*, status, headers, config*/) {
          if (data.Success) {
            resolve('reset code successfully requested');
          } else {
            reject(data.Message);
            console.log(data);
          }
        }).error(function (/*data, status, headers, config*/) {
          console.log('an error occurred');
          reject('Error connecting to MGM');
        });
      });
    };

    this.resetPassword = function (uname, token, password) {
      return $q(function (resolve, reject) {
        $http.post('/auth/passwordReset', {
          'Name': uname,
          'Token': token,
          'Password': password
        }).success(function (data /*, status, headers, config*/) {
          if (data.Success) {
            resolve('Password successfully changed');
          } else {
            reject(data.Message);
            console.log(data);
          }
        }).error(function (/*data, status, headers, config*/) {
          console.log('an error occurred');
          reject('Error connecting to MGM');
        });
      });
    };
  });
