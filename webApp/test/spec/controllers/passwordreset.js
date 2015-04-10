'use strict';

describe('Controller: PasswordresetCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var PasswordresetCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    PasswordresetCtrl = $controller('PasswordresetCtrl', {
      $scope: scope
    });
  }));

  it('should set initialize input fields', function () {
    expect(scope.input.email).toBe('');
    expect(scope.input.uname).toBe('');
    expect(scope.input.token).toBe('');
    expect(scope.input.password).toBe('');
    expect(scope.input.confirm).toBe('');
  });

  it('should set initialize error messages', function () {
    expect(scope.error.email).toBe('');
    expect(scope.error.uname).toBe('');
    expect(scope.error.token).toBe('');
    expect(scope.error.password).toBe('');
  });


});