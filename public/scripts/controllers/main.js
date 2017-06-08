'use strict';

/**
 * @ngdoc function
 * @name publicApp.controller:MainCtrl
 * @description
 * # MainCtrl
 * Controller of the publicApp
 */
angular.module('publicApp')
    .controller('MainCtrl', function ($scope, $resource) {
        var Routes = $resource('/routes', {}, {
            query: {method: 'get', isArray: true, cancellable: true}
        });
        var routes = Routes.query();
        $scope.routes = routes;
        $scope.search = "";
        $scope.searchFunc = function () {
            var newRoutes = [];
            var re = new RegExp(".*" + $scope.search + ".*");
            angular.forEach(routes, function (value, key) {
                if (!re.test(value)) {
                    return
                }
                newRoutes.push(value);
            });
            $scope.routes = newRoutes;
        };
    });
