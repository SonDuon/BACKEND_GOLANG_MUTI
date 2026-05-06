// AngularJS 1.x Application for Movie Streaming
(function() {
    'use strict';

    // Create the main application module
    angular.module('movieApp', [])
        .controller('MainController', MainController);

    // Main Controller
    MainController.$inject = ['$http', '$window'];
    
    function MainController($http, $window) {
        var vm = this;

        // API Configuration
        var API_BASE_URL = 'http://localhost:8080/api/v1';

        // View State
        vm.view = 'home'; // 'home', 'detail', 'watch', 'admin'
        vm.isAdmin = true; // Set to true to show admin panel

        // Movie List State
        vm.movies = [];
        vm.loading = false;
        vm.error = null;
        vm.currentPage = 1;
        vm.totalPages = 1;
        vm.limit = 20;
        vm.pages = [];
        vm.searchQuery = '';

        // Movie Detail State
        vm.currentMovie = null;
        vm.loadingDetail = false;

        // Watch State
        vm.watchLinks = [];
        vm.loadingWatch = false;
        vm.selectedLink = null;

        // Admin State
        vm.providers = null;
        vm.loadingAdmin = false;

        // Initialize
        vm.$onInit = function() {
            loadMovies(1);
        };

        // Load Movies List
        vm.loadMovies = function(page) {
            vm.loading = true;
            vm.error = null;
            vm.currentPage = page || 1;

            $http.get(API_BASE_URL + '/movies', {
                params: {
                    page: vm.currentPage,
                    limit: vm.limit
                }
            })
            .then(function(response) {
                var data = response.data;
                if (data.success) {
                    vm.movies = data.data || [];
                    var total = data.total || 0;
                    vm.totalPages = Math.ceil(total / vm.limit);
                    
                    // Generate pages array for pagination
                    vm.pages = [];
                    for (var i = 1; i <= vm.totalPages && i <= 10; i++) {
                        vm.pages.push(i);
                    }
                } else {
                    vm.error = 'Không thể tải danh sách phim';
                }
            })
            .catch(function(error) {
                console.error('Error loading movies:', error);
                vm.error = 'Lỗi kết nối đến server. Vui lòng kiểm tra lại.';
            })
            .finally(function() {
                vm.loading = false;
            });
        };

        // View Movie Detail
        vm.viewMovie = function(slug) {
            vm.loadMovieDetail(slug);
        };

        // Load Movie Detail
        vm.loadMovieDetail = function(slug) {
            vm.loadingDetail = true;
            vm.error = null;

            $http.get(API_BASE_URL + '/movies/' + slug)
            .then(function(response) {
                var data = response.data;
                if (data.success) {
                    vm.currentMovie = data.data;
                    vm.view = 'detail';
                } else {
                    vm.error = 'Không tìm thấy phim';
                }
            })
            .catch(function(error) {
                console.error('Error loading movie detail:', error);
                vm.error = 'Phim không tồn tại hoặc đã bị xóa';
            })
            .finally(function() {
                vm.loadingDetail = false;
            });
        };

        // Watch Movie
        vm.watchMovie = function(slug) {
            vm.loadWatchLinks(slug, 'full');
        };

        // Load Watch Links
        vm.loadWatchLinks = function(slug, episodeSlug) {
            vm.loadingWatch = true;
            vm.error = null;
            vm.selectedLink = null;

            $http.get(API_BASE_URL + '/movies/' + slug + '/watch', {
                params: {
                    episode: episodeSlug || 'full'
                }
            })
            .then(function(response) {
                var data = response.data;
                if (data.success) {
                    vm.watchLinks = data.data || [];
                    if (vm.watchLinks.length > 0) {
                        vm.selectedLink = vm.watchLinks[0].source_key;
                    }
                    vm.view = 'watch';
                } else {
                    vm.error = 'Không có link xem cho phim này';
                }
            })
            .catch(function(error) {
                console.error('Error loading watch links:', error);
                vm.error = 'Không thể tải link xem. Phim có thể chưa sẵn sàng.';
            })
            .finally(function() {
                vm.loadingWatch = false;
            });
        };

        // Select Episode
        vm.selectEpisode = function(sourceKey) {
            vm.selectedLink = sourceKey;
        };

        // Search Movies
        vm.search = function() {
            if (!vm.searchQuery || vm.searchQuery.trim() === '') {
                vm.loadMovies(1);
                return;
            }
            
            // Note: Backend cần hỗ trợ search endpoint
            // Tạm thời load tất cả và filter client-side
            vm.loading = true;
            vm.loadMovies(1).then(function() {
                // Client-side filtering (temporary solution)
                // Backend nên hỗ trợ search query parameter
            });
        };

        // Check Providers (Admin)
        vm.checkProviders = function() {
            vm.loadingAdmin = true;
            vm.error = null;

            $http.get(API_BASE_URL + '/admin/providers/health')
            .then(function(response) {
                var data = response.data;
                vm.providers = data.providers || {};
            })
            .catch(function(error) {
                console.error('Error checking providers:', error);
                vm.error = 'Không thể kiểm tra tình trạng providers';
            })
            .finally(function() {
                vm.loadingAdmin = false;
            });
        };

        // Helper: Format number with commas
        vm.formatNumber = function(num) {
            if (!num) return '0';
            return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        };

        // Initialize on load
        vm.loadMovies(1);
        
        // Check providers on admin view load
        vm.checkProviders();
    }

})();
