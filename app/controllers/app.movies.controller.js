angular
    .module('myApp')
    .controller('moviesController', function ($scope, $http) {

        $scope.titles = [];

        $scope.getTitles = function () {
            $http.get('/api/titles')
                .success(function (data) {
                    $scope.titles = data;
                }).error(function (data) {
                    console.log('Error: ' + data);
                });
        }

        $scope.clearTitles = function () {
            $scope.selTitle = "";
            $scope.titles = [];
        }
    });