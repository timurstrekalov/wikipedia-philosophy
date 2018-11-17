angular.module('wikipedia', [])
    .controller('PathController', ['$scope', '$http', function ($scope, $http) {
        var nodeIdSequence = 0;
        var edgeIdSequence = 0;
        var nodes = new vis.DataSet();
        var edges = new vis.DataSet();
        var nodeMap = {};
        var edgeMap = {};

        function Node(label, uri) {
            this.id = ++nodeIdSequence;
            this.label = label;
            this.uri = uri;
        }

        function Edge(from, to) {
            this.id = ++edgeIdSequence;
            this.from = from;
            this.to = to;
        }

        function toNode(data) {
            if (!nodeMap[data.uri]) {
                nodeMap[data.uri] = new Node(data.title, data.uri);
            }

            return nodeMap[data.uri];
        }

        function toEdge(from, to) {
            var key = from + " - " + to;
            if (!edgeMap[key]) {
                edgeMap[key] = new Edge(from, to);
            }

            return edgeMap[key];
        }

        function addPath(nodes, edges, path) {
            var newNodes = [];
            var newEdges = [];

            for (var i = 0; i < path.length - 1; i++) {
                var from = toNode(path[i]);
                var to = toNode(path[i + 1]);
                var edge = toEdge(from.id, to.id);

                newNodes.push(from);
                newNodes.push(to);
                newEdges.push(edge);
            }

            var first = newNodes[0];
            first.font = {
                color: '#ffc107',
                size: 20
            };

            var last = newNodes[newNodes.length - 1];
            last.fixed = {
                x: true,
                y: true
            };
            last.font = {
                color: '#28a745',
                size: 28
            };

            nodes.update(newNodes);
            edges.update(newEdges);
        }

        // create a network
        var container = document.getElementById('graph');
        var data = {
            nodes: nodes,
            edges: edges
        };
        var options = {
            edges: {
                arrows: {
                    to: {
                        enabled: true,
                        scaleFactor: 0.5
                    }
                }
            },
            nodes: {
                shape: 'text',
                font: {
                    size: 11
                }
            }
        };

        var network = new vis.Network(container, data, options);

        $scope.pathRequest = {
            from: 'Special:Random',
            to: 'Philosophy'
        };

        $scope.getPath = function (request) {
            $http
                .get('/api/path', {params: request})
                .then(
                    function (response) {
                        addPath(nodes, edges, response.data);
                    },
                    function (response) {
                        console.log("request failed: " + response);
                    }
                );
        };
    }]);