<template>
    <div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="true" tabindex="-1">
        <div :class="['modal-dialog', size]">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                    <h4 class="modal-title text-center text-primary"><span>{{title}}</span></h4>
                </div>
                <div class="modal-body">
                    <div class="form-inline" autocomplete="off" spellcheck="false">
                        <div class="form-group form-group-sm">
                            <label>搜索</label>
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm">
                            <label>会话类型</label>
                            <select class="form-control" v-model.trim="session_type">
                                <option value="">全部</option>
                                <option value="stream">直播</option>
                                <option value="playback">回放</option>
                                <option value="download">下载</option>
                                <option value="broadcast">对讲</option>
                            </select>
                        </div>
                        <div class="form-group form-group-sm pull-right">
                            <div class="btn-group btn-group-sm">
                                <button type="button" class="btn btn-danger" @click.prevent="clearSessions" v-if="sessions.length"><i class="fa fa-remove"></i> 清空</button>
                                <button type="button" class="btn btn-primary" @click.prevent="getSessions"><i class="fa fa-refresh"></i> 刷新</button>
                            </div>
                        </div>
                    </div>
                    <br>
                    <el-table :data="sessions" stripe :default-sort="{prop: 'StartAt', order: 'descending'}" @sort-change="sortChange" :max-height="500" v-loading="loading" element-loading-text="加载中...">
                        <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="ChannelID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column label="操作" min-width="100" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                            <template slot-scope="props">
                                <div class="btn-group btn-group-xs">
                                    <button type="button" class="btn btn-danger" @click.prevent="stopSession(props.row)">
                                        <i class="fa fa-stop"></i> 停止
                                    </button>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column prop="ChannelName" label="通道名称" min-width="120" :formatter="formatChannelName" show-overflow-tooltip></el-table-column>
                        <el-table-column prop="SessionType" label="会话类型" min-width="100" :formatter="formatSessionType"></el-table-column>
                        <el-table-column prop="StartAt" label="开始时间" min-width="160" sortable="custom"></el-table-column>
                        <el-table-column prop="Duration" label="持续时长" min-width="120" :formatter="formatSessionDuration" sortable="custom"></el-table-column>
                        <el-table-column prop="Transport" label="流传输" min-width="250" :formatter="formatTransport" show-overflow-tooltip></el-table-column>
                        <el-table-column prop="InBitRate" label="码率" min-width="120" :formatter="formatInBitRate" align="right" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="SourceVideoCodecName" label="视频编码" min-width="100" :formatter="formatVideoCodec"></el-table-column>
                        <el-table-column prop="SMS" label="SMS" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
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
    import $ from 'jquery';
    import _ from "lodash";

    export default {
        props: {
            title: {
                default: ''
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
                session_type: "",
                total: 0,
                pageSize: 10,
                currentPage: 1,
                sort: "",
                order: "",
                loading: false,
                sessions: [],
                id: '', // 外部关联 id
                serial: '', // 外部关联 serial
                name: '', // 外部关联 name
            }
        },
        watch: {
            q: function(newVal, oldVal) {
                this.doDelaySearch();
            },
            session_type: function(newVal, oldVal) {
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
                this.$emit("show");
            }).on("hidden.bs.modal", () => {
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
            formatInBitRate(row, col, cell) {
                return cell ? `${cell} Kbps` : "-";
            },
            formatVideoCodec(row, col, cell) {
                if(cell == "hevc") return "H265";
                if(cell == "h264") return "H264";
                return cell || "-";
            },
            formatTransport(row, col, cell) {
                var ret = row.CascadeTransport || "-";
                if(ret.indexOf("tcp") == 0 && row.CascadeSetup) {
                    ret += ` ${row.CascadeSetup}`;
                }
                return ret
            },
            formatChannelName(row, col, cell) {
                return cell || "-";
            },
            formatSessionType(row, col, cell) {
                switch(row.SessionType) {
                    case "stream":
                        return "直播";
                    case "playback":
                        return "回放";
                    case "download":
                        return "下载";
                    case "broadcast":
                        return "对讲";
                }
                return row.SessionType || "-";
            },
            formatSessionDuration(row, col, cell) {
                return this.formatDuration(row.Duration || 0);
            },
            getSessions(e) {
                if(!this.id) return;
                this.loading = true;
                $.get("/api/v1/cascade/session/list", {
                    id: this.id,
                    q: this.q,
                    session_type: this.session_type,
                    start: (this.currentPage -1) * this.pageSize,
                    limit: this.pageSize,
                    sort: this.sort,
                    order: this.order,
                }).then(ret => {
                    this.total = ret.total;
                    this.sessions = ret.rows || [];
                    if(e && e.target) {
                        this.$message({
                            type: "success",
                            message: "刷新成功！",
                        });
                    }
                }).always(() => {
                    this.loading = false;
                });
            },
            stopSession(row) {
                $(this.$el).hide();
                this.$confirm(`确认停止 ${row.ChannelName || row.ChannelID || row.CallID} 会话?`, "提示", {
                    lockScroll: false,
                }).then(() => {
                    $(this.$el).show().focus();
                    $.post("/api/v1/cascade/session/stop", {
                        id: this.id,
                        serial: this.serial,
                        sid: `${row.CallID}:${row.From}:${row.To}`,
                    }).always(() => {
                        this.getSessions();
                    });
                }).catch(() => {
                    $(this.$el).show().focus();
                });
            },
            clearSessions() {
                $(this.$el).hide();
                this.$confirm(`确认清空 ${this.name || this.serial} 会话?`, "提示", {
                    lockScroll: false,
                }).then(() => {
                    $(this.$el).show().focus();
                    $.post("/api/v1/cascade/session/clear", {
                        id: this.id,
                        serial: this.serial,
                    }).always(() => {
                        this.getSessions();
                    });
                }).catch(() => {
                    $(this.$el).show().focus();
                });
            },
            reset() {
                this.id = '';
                this.serial = '';
                this.name = '';
                this.sessions = [];
                this.q = "";
                this.session_type = "";
                this.currentPage = 1;
                this.pageSize = 10;
            },
            show(id, serial, name) {
                this.id = id;
                this.serial = serial || "";
                this.name = name || "";
                $(this.$el).modal("show");
                this.getSessions();
            },
            hide() {
                $(this.$el).modal("hide");
            }
        }
    }
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px){
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px){
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>
