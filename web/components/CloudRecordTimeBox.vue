<template>
<div class="box box-primary records">
    <div class="box-header">
        <h4 class="text-primary text-center">
            <a role="button" class="text-primary" @click.prevent="$refs['cloudRecordChannelListDlg'].show(day, 'timeview', serial, code)">
                云端录像({{ name }})-时间轴视图
            </a>
        </h4>
    </div>
    <div class="box-body">
        <div class="form-inline" autocomplete="off" spellcheck="false">
            <div class="form-group">
                <button type="button" class="btn btn-primary btn-sm" @click.prevent="$router.go(-1)" v-if="$getQueryString('menu') != 'no'">
                    <i class="fa fa-chevron-left"></i> 返回
                </button>
                <span v-if="shareURL">&nbsp;&nbsp;</span>
                <a role="button" class="btn btn-default btn-sm" :href="shareURL" target="_blank" v-if="shareURL">
                  <i class="fa fa-share"></i> 分享页
                </a>
                <span v-if="!isDemoUser(serverInfo, userInfo) && videos.length > 0">&nbsp;&nbsp;</span>
                <button type="button" class="btn btn-default btn-sm" @click.prevent="showTimeRangeDlg" v-if="!isDemoUser(serverInfo, userInfo) && videos.length > 0"
                    v-loading.fullscreen.lock="bDownloading" element-loading-text="合成中...">
                    <i class="fa fa-download"></i> 下载
                </button>
            </div>
            <div class="form-group pull-right">
                <div class="input-group input-group-sm">
                    <CloudRecordDatePicker class="form-control" :day="day" @update:day="updateDay" ref="datePicker" :serial="serial" :code="code"></CloudRecordDatePicker>
                    <div class="input-group-btn">
                        <button type="button" class="btn btn-sm btn-default" @click.prevent="showCloudRecordDatePicker"> <i class="fa fa-calendar"></i> </button>
                        <router-link :to="`/cloudrecord/listview/${serial}/${code}/${day}`" replace class="btn btn-default btn-sm">
                            <i class="fa fa-hand-o-right"></i> 列表视图
                        </router-link>
                    </div>
                </div>
            </div>
        </div>
        <br>
        <div class="clearfix"></div>
        <LivePlayer :videoUrl="videoUrl" :waterMark="osd" muted digitalZoom
            :hideBigPlayButton="!!serverInfo.HideBigPlayButton" :currentTime="currentTime"
            @media_info="onMediaInfo" @error="onVideoError" @ended="onVideoEnd" @timeupdate="onVideoTimeUpdate"
            v-loading="loading" :loading.sync="loading" element-loading-text="加载中..." element-loading-background="#000">
        </LivePlayer>
        <div class="text-center" v-if="isDemoUser(serverInfo, userInfo) && !bOutHevcTip">
            <br>
            提示: 演示系统限制匿名登录播放时间, 若需测试长时间播放, 请<a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html">下载使用</a>
        </div>
        <div class="text-center text-red" v-if="bOutHevcTip">
            <br>
            提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速
        </div>
        <br>
        <br>
        <CloudRecordTimeRule :videos="videos" @timeChange="onTimeChange" ref="timeRule" v-loading="loadingRecords" :minutesPerUnit="serverInfo.MinutesPerUnit||5"></CloudRecordTimeRule>
        <br>
    </div>
    <TimeRangeDlg ref="timeRangeDlg" format="HHmmss" @submit="onTimeRangeChange"></TimeRangeDlg>
    <CloudRecordChannelListDlg ref="cloudRecordChannelListDlg"></CloudRecordChannelListDlg>
</div>
</template>

<script>
import moment from 'moment'
import CloudRecordDatePicker from 'components/CloudRecordDatePicker.vue'
import CloudRecordTimeRule from 'components/CloudRecordTimeRule.vue'
import TimeRangeDlg from 'components/TimeRangeDlg.vue'
import CloudRecordChannelListDlg from "components/CloudRecordChannelListDlg.vue"
import LivePlayer from '@liveqing/liveplayer'
import { mapState } from "vuex"

export default {
    props: {
        serial: '',
        code: '',
        day: ''
    },
    data() {
        return {
            videos: [],
            videoUrl: "",
            currentTime: 0,
            bActive: false,
            loading: false,
            loadingRecords: false,
            name: '',
            osd: '',
            mediaInfo: null,
            bOutHevcTip: false,
            bDownloading: false,
            bShared: false,
        }
    },
    components: {
        CloudRecordDatePicker, LivePlayer, CloudRecordTimeRule, TimeRangeDlg, CloudRecordChannelListDlg
    },
    methods: {
        updateDay(day) {
            this.$router.replace(`/cloudrecord/timeview/${this.serial}/${this.code}/${day}`);
        },
        showCloudRecordDatePicker() {
            $(this.$refs['datePicker'].$el).focus();
        },
        updateVideos() {
            this.loadingRecords = true;
            $.get("/api/v1/cloudrecord/querydaily", {
                serial: this.serial,
                code: this.code,
                period: this.day
            }).then(data => {
                this.bShared = !!data.shared;
                this.name = data.name || `${this.serial}:${this.code}`;
                this.osd = data.osd || "";
                this.videos = data.list;
            }).always(() => {
                this.$nextTick(() => {
                    this.loadingRecords = false;
                })
            })
        },
        onTimeChange(video, bSeek) {
            this.videoUrl = (video || {}).hls || "";
            if(bSeek) {
                this.currentTime = (video || {}).currentTime || 0;
            } else {
                this.currentTime = 0;
            }
            this.bOutHevcTip = false;
        },
        onMediaInfo(mi) {
            this.mediaInfo = mi;
        },
        onVideoError(e) {
            if(e == 'MediaError' && this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) {
                if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
                    this.bOutHevcTip = true;
                    console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
                }
            }
        },
        onVideoEnd() {
            this.mediaInfo = null;
            var idx = this.videoUrls.indexOf(this.videoUrl);
            if (idx >= 0 && idx < this.videos.length - 1) {
                // this.mediaInfo = null;
                // this.bOutHevcTip = false;
                var nextVideo = this.videos[idx + 1];
                if (!nextVideo) return;
                var startTime = moment(nextVideo.startAt, 'YYYYMMDDHHmmss');
                var n = startTime.diff(startTime.clone().startOf('day'), 'minutes');
                this.$refs['timeRule'].clickMinute(n);
            }
        },
        onVideoTimeUpdate(currentTime) {
            if (currentTime < 1) return;
            // this.bOutHevcTip = false;
            var idx = this.videoUrls.indexOf(this.videoUrl);
            if (idx >= 0 && idx < this.videos.length) {
                var video = this.videos[idx];
                var startTime = moment(video.startAt, 'YYYYMMDDHHmmss');
                startTime.add(currentTime, 'seconds');
                var n = startTime.diff(startTime.clone().startOf('day'), 'minutes');
                this.$refs['timeRule'].clickMinute(n, false);
            }
        },
        showTimeRangeDlg() {
            var start = moment().second(0).subtract(30, "minutes");
            var end = moment().second(0);
            if(!start.isSame(end, "day")) {
                start = end.clone().startOf("day");
            }
            var dlg = this.$refs['timeRangeDlg'];
            dlg.setDefault([start.toDate(), end.toDate()]);
            dlg.show();
        },
        onTimeRangeChange(v) {
            if(v && v.length == 2) {
                var start = `${this.day}${v[0]}`;
                var end = `${this.day}${v[1]}`;
                var t1 = moment(start, "YYYYMMDDHHmmss");
                var t2 = moment(end, "YYYYMMDDHHmmss");
                if(t1.isBefore(t2)) {
                    var found = false;
                    for(var i = t1.clone(); i.isSameOrBefore(t2); i = i.add(1, "minute")) {
                        var secs = i.unix();
                        secs = secs - (secs % ((this.serverInfo.MinutesPerUnit||5) * 60)); // floor to unit
                        var itext = moment.unix(secs).format("HH:mm");
                        if(this.$refs["timeRule"].activeMinutes[itext]) {
                            found = true;
                            break;
                        }
                    }
                    if(!found) {
                        this.$message({
                            type: "error",
                            message: `${t1.format("HH:mm:ss")} - ${t2.format("HH:mm:ss")} 没有录像`
                        })
                        return;
                    }
                    this.$refs['timeRangeDlg'].hide();
                    this.bDownloading = true;
                    var action = t1.seconds() % 10 != 0 || t2.seconds() % 10 != 0 ? "downloadex" : "download";
                    var url = `/api/v1/cloudrecord/video/${action}/${this.serial}/${this.code}/${start}/${end}/video.mp4`;
                    $.get(url).then(ret => {
                        window.location.href = ret;
                    }).always(() => {
                        this.bDownloading = false;
                    })
                    return;
                }
            }
            this.$message({
                type: "error",
                message: `无效的时间段`
            })
        },
    },
    computed: {
        ...mapState(['userInfo', 'serverInfo']),
        shareURL() {
            if(this.bShared || this.serverInfo.APIAuth === false || this.serverInfo.DemoUser) {
                return `/playback.html?type=cloud&serial=${this.serial}&code=${this.code}&day=${this.day}`;
            }
            return "";
        },
        videoUrls() {
            return this.videos.map((val, index, array) => {
                return (val || {}).hls;
            });
        }
    },
    mounted() {
        let mmt = moment();
        let n = mmt.diff(mmt.clone().startOf('day'), 'minutes');
        n -= 10;
        if (n < 0) n = 0;
        this.$refs.timeRule.clickMinute(n);
        if (!this.serial || !this.code) {
            this.$router.replace("/cloudrecord");
            return;
        }
        if (!this.day) {
            this.$router.replace(`/cloudrecord/timeview/${this.serial}/${this.code}/${moment().format('YYYYMMDD')}`);
            return;
        }
        this.updateVideos();
    },
    beforeRouteUpdate(to, from, next) {
        if (!to.params.serial || !to.params.code) {
            next({
                path: '/cloudrecord',
                replace: true
            })
            return;
        }
        if (!to.params.day) {
            next({
                path: `/cloudrecord/timeview/${to.params.serial}/${to.params.code}/${moment().format('YYYYMMDD')}`,
                replace: true
            })
            return;
        }
        next();
        this.$nextTick(() => {
            if(!this.loadingRecords && from.path != to.path) {
                this.updateVideos();
            }
        })
    }
}
</script>

<style lang="less" scoped>
.player-wrapper {
    margin: 0 auto;
    max-width: 700px;

    @media screen and(min-width: 1600px){
        max-width: 50%;
    }
}
</style>
