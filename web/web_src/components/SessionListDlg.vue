<template>
    <div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="true" tabindex="-1">
        <div :class="['modal-dialog', size]">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                    <h4 class="modal-title text-center text-primary"><span>{{ (title&&sms) ? `${title}(${sms})` : (title||sms||'会话列表') }}</span></h4>
                </div>
                <div class="modal-body">
                    <div class="form-inline" autocomplete="off" spellcheck="false">
                        <div class="form-group form-group-sm">
                            <label>搜索</label>
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm">
                            <label>筛选</label>
                            <select class="form-control" v-model.trim="filter">
                                <option value="">全部</option>
                                <option value="stream">直播</option>
                                <option value="playback">回放</option>
                                <option value="broadcast" v-if="serverInfo.VersionType != '标准版'">对讲</option>
                                <option value="record" v-if="serverInfo.VersionType != '标准版'">录像</option>
                                <option value="cascade" v-if="serverInfo.VersionType != '标准版'">级联</option>
                                <option value="playing">播放</option>
                                <option value="udp">UDP</option>
                                <option value="tcp">TCP</option>
                                <option value="h264">H264</option>
                                <option value="hevc">H265</option>
                            </select>
                        </div>
                        <span class="hidden-xs" v-if="numOutputs && filter == 'playing'">&nbsp;&nbsp;</span>
                        <span v-if="numOutputs && filter == 'playing'">在线人数(合计): {{numOutputs}}</span>
                        <div class="form-group form-group-sm pull-right">
                            <div class="btn-group btn-group-sm">
                                <button type="button" class="btn btn-primary" @click.prevent="getSessions"><i class="fa fa-refresh"></i> 刷新</button>
                            </div>
                        </div>
                    </div>
                    <br>
                    <el-table :data="sessions" stripe @sort-change="sortChange" :max-height="500" ref="sessionTable" v-loading="loading" element-loading-text="加载中...">
                        <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="ChannelID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="ChannelName" label="通道名称" min-width="160" :formatter="formatChannelName" show-overflow-tooltip></el-table-column>
                        <el-table-column prop="SessionType" label="会话类型" min-width="100">
                            <template slot-scope="props">
                                <span :title="formatSessionTypeTitle(props.row)">{{formatSessionType(props.row)}}</span>
                            </template>
                        </el-table-column>
                        <el-table-column prop="StartAt" label="开始时间" min-width="160" sortable="custom"></el-table-column>
                        <el-table-column prop="Duration" label="持续时长" min-width="120" :formatter="formatSessionDuration" align="right" sortable="custom"></el-table-column>
                        <el-table-column prop="InBytes" label="收流大小" min-width="120" :formatter="formatInBytes" align="right" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="InBitRate" label="收流码率" min-width="120" :formatter="formatInBitRate" align="right" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="Transport" label="流传输" min-width="100" show-overflow-tooltip></el-table-column>
                        <el-table-column prop="SourceVideoCodecName" label="视频编码" min-width="100" :formatter="formatVideoCodec"></el-table-column>
                        <el-table-column prop="CloudRecord" label="云端录像" min-width="100" :formatter="formatCloudRecord" align="center" v-if="serverInfo.VersionType != '标准版'"></el-table-column>
                        <el-table-column prop="NumOutputs" label="在线人数" min-width="120" sortable="custom" :fixed="filter == 'playing' ? 'right' : false"></el-table-column>
                        <el-table-column prop="CascadeSize" label="级联数" min-width="100" sortable="custom" v-if="serverInfo.VersionType != '标准版'"></el-table-column>
                        <el-table-column prop="SMSID" label="SMS" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column label="操作" min-width="100" class-name="opt-group" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">
                            <template slot-scope="props">
                                <div class="btn-group btn-group-xs">
                                    <button type="button" class="btn btn-danger" @click.prevent="stopSession(props.row)">
                                        <i class="fa fa-stop"></i> 停止
                                    </button>
                                </div>
                            </template>
                        </el-table-column>
                    </el-table>
                    <el-pagination v-if="total > 0" layout="total,prev,pager,next,sizes" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
                    <div class="clearfix"></div>
                </div>
                <!-- <div class="modal-footer">
                    <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
                </div> -->
            </div>
        </div>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable';
    import prettyBytes from 'pretty-bytes';
    import $ from 'jquery';
    import _ from "lodash";

    export default {
        props: {
            title: {
                default: '会话列表'
            },
            size: {
                type: String,
                default: 'modal-lgg'
            },
            fade: {
                type: Boolean,
                default: false
            },
            serverInfo: {
                type: Object,
                default: () => {}
            },
            userInfo: {
                type: Object,
                default: () => null
            }
        },
        data() {
            return {
                q: "",
                filter: "",
                sms: "",
                total: 0,
                pageSize: 10,
                currentPage: 1,
                sort: "StartAt",
                order: "",
                loading: false,
                shown: false,
                numOutputs: 0,
                sessions: [],
            }
        },
        watch: {
            q: function(newVal, oldVal) {
                this.doDelaySearch();
            },
            filter: function(newVal, oldVal) {
                this.doSearch();
            },
            currentPage: function(newVal, oldVal) {
                this.doSearch(newVal);
            },
            pageSize: function(newVal, oldVal) {
                this.doSearch();
            }
        },
        mounted() {
            $(this.$el).find('.modal-content').draggable({
                handle: '.modal-header',
                cancel: ".modal-title span",
                addClasses: false,
                containment: 'document',
                delay: 100,
                opacity: 0.5
            });
            $(this.$el).on("shown.bs.modal", () => {
                this.shown = true;
                this.$emit("show");
                this.$refs["sessionTable"].sort("StartAt", "descending");
            }).on("hidden.bs.modal", () => {
                this.shown = false;
                this.errors.clear();
                this.reset();
                this.$emit("hide");
            })
        },
        methods: {
            sortChange(data) {
                this.sort = data.prop;
                this.order = data.order == "ascending" ? "asc" : "desc";
                this.getSessions();
            },
            doSearch(page = 1) {
                this.currentPage = page;
                this.getSessions();
            },
            doDelaySearch: _.debounce(function() {
                this.doSearch();
            }, 800),
            formatChannelName(row, col, cell) {
                return cell || "-";
            },
            formatSessionType(row, col, cell) {
                var strs = (row.StreamID || "").split(":");
                if(strs.length > 1) {
                    switch(strs[0]) {
                        case "stream":
                            return "直播";
                        case "playback":
                            return "回放";
                        case "download":
                            if(row.CloudRecord) {
                                return "补录";
                            } else {
                                return "下载";
                            }
                        case "broadcast":
                            return "对讲";
                        default:
                            return strs[0] || "-";
                    }
                }
                return "-";
            },
            formatSessionTypeTitle(row, col, cell) {
                if(row.PlaybackStartTime && row.PlaybackEndTime) {
                    return `${row.PlaybackStartTime} ~ ${row.PlaybackEndTime}`;
                }
                return "";
            },
            formatSessionDuration(row, col, cell) {
                return this.formatDuration(row.Duration || 0);
            },
            formatInBytes(row, col, cell) {
                return cell ? prettyBytes(cell) : "-";
            },
            formatInBitRate(row, col, cell) {
                return cell ? `${cell} Kbps` : "-";
            },
            formatVideoCodec(row, col, cell) {
                if(cell == "hevc") return "H265";
                if(cell == "h264") return "H264";
                return cell || "-";
            },
            formatCloudRecord(row, col, cell) {
                return cell ? "✓" : "✗";
            },
            getSessions(e) {
                if(!this.shown) return;
                this.loading = true;
                $.get("/api/v1/device/session/list", {
                    q: this.q,
                    filter: this.filter,
                    sms: this.sms,
                    start: (this.currentPage -1) * this.pageSize,
                    limit: this.pageSize,
                    sort: this.sort,
                    order: this.order
                }).then(ret => {
                    this.numOutputs = ret.NumOutputs || 0;
                    this.total = ret.SessionCount || 0;
                    this.sessions = ret.SessionList || [];
                    if(e && e.target) {
                        this.$message({
                            type: "success",
                            message: "刷新成功！",
                        });
                    }
                }).always(() => {
                    this.$nextTick(() => {
                        this.loading = false;
                    })
                });
            },
            stopSession(row) {
                $(this.$el).hide();
                this.$confirm(`确认停止 ${row.ChannelName || row.ChannelID} 会话?`, "提示", {
                    lockScroll: false,
                }).then(() => {
                    $(this.$el).show().focus();
                    $.post("/api/v1/device/session/stop", {
                        streamid: row.StreamID,
                    }).always(() => {
                        this.getSessions();
                    });
                }).catch(() => {
                    $(this.$el).show().focus();
                });
            },
            reset() {
                this.total = 0;
                this.numOutputs = 0;
                this.sessions = [];
                this.q = "";
                this.filter = "";
                this.sms = "";
                this.currentPage = 1;
                this.pageSize = 10;
                this.$refs["sessionTable"].clearSort();
            },
            show(filter = "", sms = "") {
                this.filter = filter;
                this.sms = sms;
                this.$nextTick(() => {
                    $(this.$el).modal("show");
                });
            },
            hide() {
                $(this.$el).modal("hide");
            },
        }, //-- methods
    }
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px) {
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px) {
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>
