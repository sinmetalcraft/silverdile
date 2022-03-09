# silverdile

Google App Engine Image Serviceのようなものを2nd genでも利用するために生まれた

## アーキテクチャ

### v1

変換要求が来る度に画像を生成し、エッジキャッシュに乗せて返す

### v2

変換要求が来た時に、変換後ObjectBucketを参照し、そこにすでに画像があれば、それを返す。
画像がない場合、画像を生成し、変換後ObjectBucketに保存して、それを返す。

### v3

アーキテクチャはv2と同様。
singleflightやcache layerを独自に入れやすいようにv2と比べて処理が分解されて、必要最低限のものだけになった。
Cloud Runで動かすことを意識しているが、GAE 2nd Genでも動く。

## Setup

オリジナルの画像が入ったBucketNameにPrefixとして `alter-` を付けたBucketを作成する。
`alter-` Bucketに必要であれば [Lifecycle](https://cloud.google.com/storage/docs/lifecycle) を設定して、変換後イメージの削除を行う。

## 機能

### Resize

Pathとして `/{bucket}/{object}/=s???` を指定すると画像の長辺を???のサイズにした状態で画像を返す

Example

* https://silverdile-dot-sinmetal-ci.an.r.appspot.com/v2/image/resize/sinmetal-ci-silverdile/jun0.jpg/=s700
* https://silverdile-dot-sinmetal-ci.an.r.appspot.com/v3/image/resize/sinmetal-ci-silverdile/jun0.jpg/=s700

#### Limitation

* Size 0 ~ 2560

## Dev

### Local GAE Run

`go run github.com/sinmetalcraft/silverdile/app`

### Deploy

`gcloud app deploy app/app.yaml`