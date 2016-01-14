'use strict';
var gameStatus;

// Declare app level module which depends on views, and components
angular.module('DrinkEm', [
  'ngWebSocket',
  'ngMaterial',
]).
factory('GameData', ['$websocket', function($websocket) {
  // Open a WebSocket connection
  var dataStream = $websocket('ws://' + host + '/data');

  return {
    start: function(callback) {
      dataStream.onMessage(function(message) {
        var gameState = Game.StateFromBlob(message.data);
        gameStatus = gameState;
        callback(gameState)
      });
    }
  }
}]).  
controller('GameCtrl', ['$scope', '$http', '$mdDialog', 'GameData', function($scope, $http, $mdDialog, GameData) {
  setScope(initialGameState);

  GameData.start(function (gameState) {
    $scope.$apply(function () {
      setScope(gameState);
    });
  });

  function setScope(gameData) {
    $scope.endgame = gameData.EndGameJS();
    if ($scope.endgame.length > 0) {
      $scope.foldOut = gameData.WinByFoldJS();
      $scope.phs = gameData.PlayerHandPathsJS();
      $scope.dc = gameData.DrawnCardPathsJS();
      $scope.community = gameData.CommunityPathsJS();
      $scope.descs = gameData.HandDescsJS();
      $scope.popprob = gameData.PopProbJS();
    } else {
      $scope.endgame = null;
      var me = gameData.MeJS();
      var players = gameData.PlayersJS();
      var activePlayerID = gameData.ActivePlayerIDJS();
      var community = gameData.CommunityPathsJS();
      for (var i = 0; i < players.length; i++) {
        if (players[i].ID === me.ID) {
          // Even if it doesn't show in the data, we know we're in
          players[i].LoggedIn = true;
        }
        if (players[i].ID === activePlayerID) {
          $scope.activePlayer = players[i];
        }
      }
      $scope.players = players;
      $scope.me = me;
      $scope.started = gameData.Started();
      $scope.gameover = gameData.GameOver();
      $scope.pot = gameData.TotalPot();
      $scope.events = gameData.RoundHistoryJS();
      $scope.hand = gameData.HandPathsJS();
      $scope.community = community;
      $scope.largeBlindID = gameData.LargeBlindIDJS();
      $scope.smallBlindID = gameData.SmallBlindIDJS();
      $scope.activePlayerID = activePlayerID;
      $scope.toPlay = gameData.StayInJS()-me.Pot;
      $scope.minBet = gameData.MinimumBetJS();
      $scope.bet = gameData.MinimumBetJS();
      $scope.eliminated = gameData.EliminatedMapJS();
      $scope.act = gameData.ActMapJS();

      // The value that determines if the player is betting. Zero if we're past
      // the first round, largeBlind otherwise
      $scope.betStr = gameData.IsBet() ? "Bet" : "Raise";
      $scope.checkStr = (gameData.StayInJS()-me.Pot) === 0 ? "Check" : "Call";
    }
  }

  function showDialog() {
    $mdDialog.show({
      templateUrl: 'game/dialog.html',
      parent: angular.element(document.body),
      escapeToClose: false,
      fullscreen: true,
      scope: $scope
    });
  }

  $scope.fold = function() {
    $http.post('/fold');
  };

  $scope.check = function() {
    $http.post('/check');
  };

  $scope.placeBet = function(bet) {
    $http.post('/bet', {bet: bet});
  };

  $scope.inc = function(amt) {
    amt = Math.min(amt, $scope.me.Money - $scope.bet);
    $scope.bet += amt;
  };

  $scope.dec = function(amt) {
    amt = Math.min(amt, $scope.bet - $scope.minBet);
    $scope.bet -= amt;
  };
}]).
filter('loggedin', function() {
   return function(items) {
    var filtered = [];

    angular.forEach(items, function(item) {
      if (item.LoggedIn) {
        filtered.push(item);
      }
    });

    return filtered;
  };
}).
directive('player', function() {
  return {
    restrict: 'E',
    scope: {
      player: '=',
      me: '=',
      eliminated: '=',
      smallBlind: '=',
      largeBlind: '=',
      activePlayer: '='
    },
    templateUrl: 'game/player.html'
  };
}).
directive('slide', function($timeout) {
  return {
    restrict: 'A',
    link: function(scope, element, attrs) {
      scope.$watch(attrs.slide, function(newValue, oldValue) {
        $timeout(function() {
          var scrollTo = $(".player-" + newValue + "-add, .player-" + newValue);
          var element = $('.player-holder');
          if (scrollTo.length) {
            element.scrollLeft(0)
            element.animate({
              scrollLeft: scrollTo.offset().left}, 500);
          }
        });
      });
    }
  };
}).
config(function($mdIconProvider) {
    $mdIconProvider
      .icon("minus", "bower_components/material-design-icons/content/svg/production/ic_remove_48px.svg", 48)
      .icon("plus", "bower_components/material-design-icons/content/svg/production/ic_add_48px.svg", 48)
      .icon("you", "bower_components/material-design-icons/action/svg/production/ic_account_circle_48px.svg", 48)
      .icon("waiting", "bower_components/material-design-icons/action/svg/production/ic_alarm_48px.svg", 48)
      .icon("small-blind", "bower_components/material-design-icons/maps/svg/production/ic_local_atm_48px.svg", 48)
      .icon("large-blind", "bower_components/material-design-icons/editor/svg/production/ic_attach_money_48px.svg", 48)
      .icon("thinking", "bower_components/material-design-icons/action/svg/production/ic_pan_tool_48px.svg", 48)
      .icon("folded", "bower_components/material-design-icons/action/svg/production/ic_highlight_off_48px.svg", 48)
      .icon("dead", "bower_components/material-design-icons/social/svg/production/ic_mood_bad_48px.svg", 48);
});
