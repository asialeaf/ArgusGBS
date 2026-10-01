<template>
<div class="box box-primary videos">
    <div class="box-header">
        <h4 class="text-primary text-center">
            <a role="button" class="text-primary" @click.prevent="$refs['cloudRecordChannelListDlg'].show(day, 'listview', serial, code)">
                云端录像({{ name }})-列表视图
            </a>
        </h4>
    </div>
    <div class="box-body">
        <div class="form-inline" autocomplete="off" spellcheck="false">
            <div class="form-group">
                <div class="btn-group">
                    <button type="button" class="btn btn-primary btn-sm" @click.prevent="$router.go(-1)" v-if="$getQueryString('menu') != 'no'">
                        <i class="fa fa-chevron-left"></i> 返回
                    </button>
                    <button type="button" class="btn btn-danger btn-sm" @click.prevent="removeDaily" v-if="videos.length > 0 && hasAnyRole(serverInfo, userInfo, '管理员')">
                        <i class="fa fa-remove"></i> 删除当天
                    </button>
                </div>
            </div>
            <div class="form-group pull-right">
                <div class="input-group input-group-sm">
                    <CloudRecordDatePicker class="form-control" :day="day" @update:day="updateDay" ref="datePicker" :serial="serial" :code="code"></CloudRecordDatePicker>
                    <div class="input-group-btn">
                        <button type="button" class="btn btn-sm btn-default" @click.prevent="showCloudRecordDatePicker">
                            <i class="fa fa-calendar"></i>
                        </button>
                        <router-link :to="`/cloudrecord/timeview/${serial}/${code}/${day}`" replace class="btn btn-default btn-sm">
                            <i class="fa fa-hand-o-right"></i> 时间轴视图
                        </router-link>
                    </div>
                </div>
            </div>
        </div>
        <br>
        <div class="clearfix"></div>
        <el-table :row-class-name="tableRowClassName" v-loading="loading" element-loading-text="加载中..."
            ref="recordTable" :data="pageData" empty-text="暂无数据, 请另选日期" :default-sort="{prop: 'startAt', order: 'descending'}" @sort-change="sortChange">
            <el-table-column min-width="200" label="设备国标编号" prop="serial" show-overflow-tooltip></el-table-column>
            <el-table-column min-width="200" label="通道国标编号" prop="code" show-overflow-tooltip></el-table-column>
            <el-table-column min-width="120" label="通道名称" prop="name" show-overflow-tooltip></el-table-column>
            <el-table-column min-width="190" label="操作" :fixed="isMobile() ? false : 'right'">
                <template slot-scope="scope">
                    <div class="btn-group">
                        <a role="button" class="btn btn-primary btn-xs" @click.prevent="play(scope.row)">
                            <i class='fa fa-play'></i> 播放
                        </a>
                        <a role="button" class="btn btn-info btn-xs" @click.prevent="download(scope.row)" v-loading.fullscreen.lock="bDownloading" element-loading-text="合成中...">
                            <i class='fa fa-download'></i> 下载
                        </a>
                        <a role="button" class="btn btn-danger btn-xs" @click.prevent="remove(scope.row)" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                            <i class="fa fa-remove"></i> 删除
                        </a>
                    </div>
                </template>
            </el-table-column>
            <el-table-column min-width="160" label="开始时间" prop="startAt" sortable="custom"></el-table-column>
            <el-table-column min-width="120" label="录像时长" prop="duration" sortable="custom"></el-table-column>
            <el-table-column min-width="100" label="文件大小" prop="filesize" :formatter="formatFileSize" align="right" sortable="custom"></el-table-column>
            <el-table-column prop="important" label="重要标记" min-width="100" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                <template slot-scope="scope">
                    <el-switch v-model="scope.row.important" @change="turnImportant(scope.row)"></el-switch>
                </template>
            </el-table-column>
            <el-table-column min-width="450" label="视频地址">
                <template slot-scope="scope">
                    <span class="ellipsis">
                        <i title="点击拷贝" role="button" v-clipboard="scope.row.hls" @success="$message({type:'success', message:'成功拷贝到粘贴板'})" class="fa fa-copy"></i>
                        <a role="button" class="text-primary" @click.prevent="play(scope.row)">{{scope.row.hls}}</a>
                    </span>
                </template>
            </el-table-column>
        </el-table>
    </div>
    <div class="box-footer clearfix" v-if="total > 0">
        <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
    </div>

    <CloudRecordVideoDlg ref="cloudRecordVideoDlg" :serverInfo="serverInfo" :userInfo="userInfo"></CloudRecordVideoDlg>
    <CloudRecordChannelListDlg ref="cloudRecordChannelListDlg"></CloudRecordChannelListDlg>
</div>
</template>

<script>
import { mapState } from "vuex";
import _ from "lodash";
import moment from "moment";
import prettyBytes from 'pretty-bytes';
import CloudRecordVideoDlg from "components/CloudRecordVideoDlg.vue";
import CloudRecordDatePicker from "components/CloudRecordDatePicker.vue";
import CloudRecordChannelListDlg from "components/CloudRecordChannelListDlg.vue";

export default {
    props: {
        serial: "",
        code: "",
        day: ""
    },
    data() {
        return {
            name: "",
            currentPage: 1,
            pageSize: 10,
            sort: "startAt",
            order: "ascending",
            loading: false,
            bDownloading: false,
            videos: [],
            bgcolor: "",
            pathname: location.pathname == "/" ? "" : location.pathname.substring(0, location.pathname.length - 1)
        };
    },
    components: {
        CloudRecordDatePicker, CloudRecordVideoDlg, CloudRecordChannelListDlg
    },
    methods: {
        tableRowClassName({row, rowIndex}) {
            if (row.important) {
                return "warning-row";
            }
        },
        isMobile() {
            return videojs.browser.IS_IOS || videojs.browser.IS_ANDROID;
        },
        updateDay(day) {
            this.$router.replace(`/cloudrecord/listview/${this.serial}/${this.code}/${day}`);
        },
        showCloudRecordDatePicker() {
            $(this.$refs.datePicker.$el).focus();
        },
        updateVideos() {
            this.loading = true;
            $.get("/api/v1/cloudrecord/querydaily", {
                serial: this.serial,
                code: this.code,
                period: this.day,
                sort: this.sort,
                order: this.order
            }).then(data => {
                this.name = data.name || `${this.serial}:${this.code}`;
                this.osd = data.osd || "";
                this.videos = data.list;
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
            })
        },
        updateVideosDebounce: _.debounce(function () {
            this.updateVideos()
        }, 800),
        play(row) {
            this.$refs["cloudRecordVideoDlg"].play(
                row.hls,
                `${this.name}(${row.startAt})`,
                row.Snap || "",
                this.osd
            );
        },
        download(row) {
            this.bDownloading = true;
            var url = `/api/v1/cloudrecord/download/${this.serial}/${this.code}/${row._startAt}`;
            $.get(url).then(ret => {
                window.location.href = ret;
            }).always(() => {
                this.bDownloading = false;
            })
        },
        removeDaily() {
            this.$confirm(`确认删除 ${this.name} ${moment(this.day).format("YYYY-MM-DD")} 当天所有录像?`,"提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/cloudrecord/removedaily", {
                    serial: this.serial,
                    code: this.code,
                    period: this.day
                }).always(() => {
                    this.updateVideos();
                    this.$refs["datePicker"].update();
                });
            }).catch(() => {});
        },
        remove(row) {
            this.$confirm(`确认删除 ${row.name} ?`, "提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/cloudrecord/remove", {
                    serial: this.serial,
                    code: this.code,
                    period: row._startAt
                }).always(() => {
                    this.updateVideos();
                });
            }).catch(() => {});
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order;
            this.updateVideos();
        },
        turnImportant(row) {
            $.post("/api/v1/cloudrecord/setimportant", {
                serial: row.serial,
                code: row.code,
                period: row._startAt,
                important: row.important
            }).then(() => {
                var msg = row.important ? "标记成功,将不会被自动清理" : "取消标记成功";
                this.$message({
                    type: "success",
                    message: msg
                });
            }).fail(() => {
                this.$set(row, "important", !row.important);
            })
        },
        formatFileSize(row, col, cell) {
            return cell ? prettyBytes(cell) : "-";
        },
    },
    mounted() {
        if (!this.serial || !this.code) {
            this.$router.replace("/cloudrecord");
            return;
        }
        if (!this.day) {
            this.$router.replace(`/cloudrecord/${this.serial}/${this.code}/${moment().format("YYYYMMDD")}`);
            return;
        }
        // this.updateVideos(); // load when sort change
    },
    computed: {
        ...mapState(["serverInfo", "userInfo"]),
        total() {
            return this.videos.length;
        },
        pageData() {
            let start = (this.currentPage - 1) * this.pageSize;
            let end = start + this.pageSize;
            return this.videos.slice(start, end).map(row => {
                if (!row.serial) {
                    row.serial = this.serial,
                    row.code = this.code,
                    row.name = this.name,
                    row.location = location;
                    row._startAt = row.startAt;
                    row.startAt = moment(row.startAt, "YYYYMMDDHHmmss").format("YYYY-MM-DD HH:mm:ss");
                    row.duration = moment.utc((row.duration || 0) * 1000).format("HH:mm:ss");
                }
                return row;
            });
        }
    },
    beforeRouteUpdate(to, from, next) {
        if (!to.params.serial || !to.params.code) {
            this.$router.replace(`/cloudrecord`);
            return;
        }
        if (!to.params.day) {
            this.$router.replace(`/cloudrecord/${to.params.serial}/${to.params.code}/${moment().format("YYYYMMDD")}`);
            return;
        }
        next();
        this.$nextTick(() => {
            if(!this.loading && from.path != to.path) {
                this.updateVideos();
            }
        });
    }
};
</script>

<style>
.el-table .warning-row {
    background: oldlace;
    color: red;
}

.el-table .success-row {
    background: #f0f9eb;
}
</style>
