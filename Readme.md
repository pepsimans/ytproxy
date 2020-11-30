        ## Description

I own WDTV media player, with installed [wdlxtv](http://forum.wdlxtv.com) firmware.
It have youtube3 plugin, which is slow and sometimes getting broken, due changes which are breaking playback functionality.

So that I've wrapped youtube proxy tool into Go microservice. This project can be easily deployed using free tier Heroku.

Getting video's direct link using this way takes ~5 seconds, and this is ok for for my purposes.
Original Youtube proxy script have performed 3 requests to the Youtube site, and it in total took about 15-20 seconds, since there was not any cache.

#### How to start
1. Git clone: `git clone https://github.com/ebogdanov/ytproxy.git`

2. Install Heroku
`https://devcenter.heroku.com/articles/heroku-cli#download-and-install`

3. Create project in Heroku
`heroku create --manifest`

4. Deploy project
`git push heroku main`

5. Script can be executed via GET path
`http://<your-heroku-project-name>/get_url?id=<youtubeVideoId>`

#### Additional settings
1. If you need authorization this can be set via environment settings SERIAL_NO. This can be setup via [Settings -> Config Vars](https://dashboard.heroku.com/apps/).
2. In case of problems you can view application logs via command `heroku logs --tail`

