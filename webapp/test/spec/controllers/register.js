'use strict';

describe('Controller: RegisterCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var RegisterCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    RegisterCtrl = $controller('RegisterCtrl', {
      $scope: scope
    });
  }));

  it('should set default state to Eupa', function () {
    expect(scope.stage).toBe('Eupa');
  });

  it('should transition to acount page', function () {
    expect(scope.stage).toBe('Eupa');
    scope.acceptEupa();
    expect(scope.stage).toBe('Account');
  });

  it('should set initialize error messages', function () {
    expect(scope.error.fname).toBe('');
    expect(scope.error.lname).toBe('');
    expect(scope.error.email).toBe('');
    expect(scope.error.pword).toBe('');
    expect(scope.error.gender).toBe('');
    expect(scope.error.other).toBe('');
  });

  it('should set clear error messages', function () {
    scope.error.fname = 'test';
    scope.error.lname = 'test';
    scope.error.email = 'test';
    scope.error.pword = 'test';
    scope.error.gender = 'test';
    scope.error.other = 'test';
    scope.resetErrors();
    expect(scope.error.fname).toBe('');
    expect(scope.error.lname).toBe('');
    expect(scope.error.email).toBe('');
    expect(scope.error.pword).toBe('');
    expect(scope.error.gender).toBe('');
    expect(scope.error.other).toBe('');
  });

  it('should set initialize registration fields', function () {
    expect(scope.registrant.fname).toBe('');
    expect(scope.registrant.lname).toBe('');
    expect(scope.registrant.email).toBe('');
    expect(scope.registrant.pword).toBe('');
    expect(scope.registrant.pword2).toBe('');
    expect(scope.registrant.gender).toBe('');
    expect(scope.registrant.reasons).toBe('');
  });
});