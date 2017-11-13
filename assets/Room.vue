<div class="flexbox-wrapper">
  <div class="flexbox-row queue-holder">
    <div class="queue">
      {{ template "queue" .}}
    </div>
  </div>
  <div class="flexbox-row flexbox-row-fill">
    <div class="container">
      <form class="search-form" action="/rooms/{{ .Room.ID }}/search">
        <div class="row">
          <div class="seven columns offset-by-two">
            <input type="text" name="search" class="u-full-width" placeholder="Search for Music">
          </div>
          <div>
            <button type="submit" class="two columns button button-primary search">Search</button>
          </div>
        </div>
      </form>
    </div>
    <div class="results"></div>
  </div>
  <div class="flexbox-row flexbox-row-fixed now-playing">
    {{ template "playing" .Tracks }}
  </div>
</div>
