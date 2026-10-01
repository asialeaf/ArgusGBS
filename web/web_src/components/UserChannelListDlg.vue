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
                            <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员', '管理员')">
                                <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                                <div class="input-group-btn">
                                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出用户通道" :disabled="hasAllChannel"><i class="fa fa-download"></i></button>
                                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="upload" title="导入用户通道" :disabled="hasAllChannel"><i class="fa fa-upload"></i></button>
                                </div>
                            </div>
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm">
                            <label>通道类型</label>
                            <select class="form-control" v-model.trim="channel_type">
                                <option value="">全部</option>
                                <option value="device">设备</option>
                                <option value="decode">解码器</option>
                                <option value="dir">子目录</option>
                            </select>
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm" v-if="channel_type == 'device' || channel_type == 'decode'">
                            <label>在线状态</label>
                            <select class="form-control" v-model.trim="online">
                                <option value="">全部</option>
                                <option value="true">在线</option>
                                <option value="false">离线</option>
                            </select>
                        </div>
                        <span class="hidden-xs" v-if="channel_type == 'device' || channel_type == 'decode'">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm">
                            <div class="checkbox" style="user-select:none;">
                                <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model.trim="related" name="Related" :disabled="hasAllChannel">
                                    只看{{reverse ? '未选' : '已选'}}({{relateCnt}})
                                </el-checkbox>
                                <span :style="hasAllChannel ? 'cursor:not-allowed;' : ''">
                                    <a role="button" @click="toggleReverse" :style="{'pointer-events': hasAllChannel ? 'none' : 'all'}" :class="{'text-gray': hasAllChannel}">
                                        <i class="fa fa-sort"></i>
                                    </a>
                                </span>
                            </div>
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm" v-if="!userInfo || userInfo.HasAllChannel">
                            <div class="checkbox">
                                <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model="hasAllChannel" @change="toggleHasAllChannel" name="HasAllChannel">全部关联</el-checkbox>
                            </div>
                        </div>
                    </div>
                    <br>
                    <el-table :data="channels" stripe @sort-change="sortChange" @select="select" @select-all="selectAll" :max-height="500"
                        ref="channelTable" v-loading="loading" element-loading-text="加载中...">
                        <el-table-column type="selection" width="55" fixed :selectable="selectable"></el-table-column>
                        <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="ID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <!-- <el-table-column prop="DeviceName" label="设备名称" min-width="120" :formatter="formatDeviceName" show-overflow-tooltip></el-table-column> -->
                        <el-table-column prop="Name" label="通道名称" min-width="160" :formatter="formatName" show-overflow-tooltip></el-table-column>
                        <el-table-column min-width="100" label="快照">
                            <template slot-scope="props">
                                <span v-if="isDir(props.row)">
                                    <i class="fa fa-sitemap"></i>
                                </span>
                                <el-popover v-else :open-delay="1000" :close-delay="10" placement="left" :title="props.row.ID" width="400" trigger="hover">
                                    <img onerror="this.src='/images/default_snap.png';" style="width:100%;height:100%;" :src="props.row.SnapURL">
                                    <img onerror="this.src='/images/default_snap.png';" style="height:30px;width:50px;" slot="reference" :src="props.row.SnapURL">
                                </el-popover>
                            </template>
                        </el-table-column>
                        <el-table-column prop="Status" label="在线状态" min-width="100">
                            <template slot-scope="props">
                                <span v-if="isDir(props.row)">-</span>
                                <span v-else-if="props.row.DeviceOnline && (props.row.CustomStatus || props.row.Status) == 'ON'" :class="{'text-success': props.row.Status == 'ON', 'text-orange': !!props.row.CustomStatus}">在线</span>
                                <span v-else :class="{'text-orange': !!props.row.CustomStatus}">离线</span>
                            </template>
                        </el-table-column>
                        <!-- <el-table-column prop="SubCount" label="子节点数" min-width="100" sortable="custom"></el-table-column> -->
                        <el-table-column prop="Manufacturer" label="厂家" min-width="120" :formatter="formatManufacturer" show-overflow-tooltip></el-table-column>
                    </el-table>
                    <el-pagination v-if="total > 0" layout="total,prev,pager,next,sizes" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
                    <div class="clearfix"></div>
                </div>
                <!-- <div class="modal-footer">
                    <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
                </div> -->
            </div>
        </div>
        <el-upload :action="`/api/v1/user/channel/import?id=${id}`" accept=".xlsx" class="hide"
            :show-file-list="false"
            :on-success="onUploadSuccess" :on-error="onUploadError" :on-progress="onUploadProgress">
            <a role="button" title="上传用户通道" ref="uploadButton">
                <i class="fa fa-upload"></i> 上传用户通道
            </a>
        </el-upload>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable'
    import $ from 'jquery'
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
                channel_type: "",
                online: "",
                total: 0,
                relateCnt: 0,
                pageSize: 10,
                currentPage: 1,
                sort: "",
                order: "",
                related: false,
                reverse: false,
                hasAllChannel: false,
                loading: false,
                channels: [],
                selection: [],
                id: '', //外部关联id
            }
        },
        watch: {
            q: function(newVal, oldVal) {
                this.doDelaySearch();
            },
            channel_type: function(newVal, oldVal) {
                this.doSearch();
            },
            online: function(newVal, oldVal) {
                this.doSearch();
            },
            related: function(newVal, oldVal) {
                this.doSearch();
            },
            reverse: function(newVal, oldVal) {
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
                this.getChannels();
            },
            select(selection, row) {
                var idx = selection.indexOf(row);
                if(idx >= 0) {
                    $.post("/api/v1/user/savechannels", {
                        id: this.id,
                        channels: [`${row.DeviceID}:${row.ID}`]
                    }).always(() => {
                        this.getChannels();
                    })
                } else {
                    $.post("/api/v1/user/removechannels", {
                        id: this.id,
                        channels: [`${row.DeviceID}:${row.ID}`]
                    }).always(() => {
                        this.getChannels();
                    })
                }
            },
            selectAll(selection) {
                if(this.hasAllChannel) return;
                var keys = [];
                if(selection.length) {
                    for(var row of selection) {
                        var idx = this.selection.indexOf(row);
                        if(idx < 0) {
                            keys.push(`${row.DeviceID}:${row.ID}`)
                        }
                    }
                    $.post("/api/v1/user/savechannels", {
                        id: this.id,
                        channels: keys,
                    }).always(() => {
                        this.getChannels();
                    })
                } else {
                    for(var row of this.selection) {
                        keys.push(`${row.DeviceID}:${row.ID}`)
                    }
                    $.post("/api/v1/user/removechannels", {
                        id: this.id,
                        channels: keys,
                    }).always(() => {
                        this.getChannels();
                    })
                }
            },
            doSearch(page = 1) {
                this.currentPage = page;
                this.getChannels();
            },
            doDelaySearch: _.debounce(function() {
                this.doSearch();
            }, 800),
            formatName(row, col, cell) {
                var devName = row.DeviceCustomName || row.DeviceName || "";
                var chName = row.CustomName || row.Name || "";
                if(devName && devName != chName) {
                    if(!chName) {
                        chName = devName;
                    } else {
                        chName = `${chName}@${devName}`;
                    }
                }
                return chName || "-";
            },
            formatDeviceName(row, col, cell) {
                return row.DeviceCustomName || row.DeviceName || "-";
            },
            formatChannelName(row, col, cell) {
                return row.CustomName || row.Name || "-";
            },
            formatManufacturer(row, col, cell) {
                if (cell) return cell;
                return "-";
            },
            selectable(row, index) {
                if(!this.hasAllChannel) {
                    return true;
                }
                return false;
            },
            getChannels() {
                if(!this.id) return;
                this.loading = true;
                $.get("/api/v1/user/channellist", {
                    id: this.id,
                    q: this.q,
                    start: (this.currentPage -1) * this.pageSize,
                    limit: this.pageSize,
                    channel_type: this.channel_type,
                    online: (this.channel_type == 'device' || this.channel_type == 'decode') ? this.online : '',
                    related: this.related,
                    reverse: this.reverse,
                    sort: this.sort,
                    order: this.order
                }).then(ret => {
                    this.$refs["channelTable"].clearSelection();
                    this.total = ret.ChannelCount;
                    this.relateCnt = ret.ChannelRelateCount;
                    this.hasAllChannel = !!ret.HasAllChannel;
                    this.channels = ret.ChannelList || [];
                    this.selection = [];
                    this.$nextTick(() => {
                        this.channels.forEach(row => {
                            var sel = row.UserID !== 0;
                            this.$refs["channelTable"].toggleRowSelection(row, sel);
                            if(sel) {
                                this.selection.push(row);
                            }
                        })
                    })
                }).always(() => {
                    this.$nextTick(() => {
                        this.loading = false;
                    })
                });
            },
            isDir(row) {
                // return row && (row.SubCount > 0 || row.Parental == 1);
                if (row) {
                    if (row.SubCount > 0) return true;
                    if (this.serverInfo.StrictChannelParental && row.Parental == 1) return true;
                    if (this.serverInfo.StrictChannelCode && row.ID.length <= 10) return true;
                    if (row.Parental == 1 && row.Manufacturer == "LiveQing") return true;
                    // if (row.ID.length == 20 && (row.ID.substring(10, 13) == "216" || row.ID.substring(10, 13) == "215" || row.ID.substring(10, 13) == "200")) {
                    if (row.ID.length == 20 && (row.ID.substring(10, 13) == "216" || row.ID.substring(10, 13) == "215")) {
                        return true;
                    }
                }
                return false;
            },
            reset() {
                this.id = '';
                this.$refs["channelTable"].clearSelection();
                this.channels = [];
                this.selection = [];
                this.q = "";
                this.channel_type = "";
                this.online = "";
                this.related = false;
                this.reverse = false;
                this.total = 0;
                this.relateCnt = 0;
                this.hasAllChannel = false;
                this.currentPage = 1;
                this.pageSize = 10;
            },
            toggleHasAllChannel(val) {
                $.post("/api/v1/user/sethasallchannel", {
                    id: this.id,
                    hasallchannel: val,
                }).always(() => {
                    this.doSearch();
                })
            },
            show(id) {
                this.id = id;
                $(this.$el).modal("show");
                this.getChannels();
            },
            hide() {
                $(this.$el).modal("hide");
            },
            download() {
                let link = `/api/v1/user/channel/export?id=${this.id}`;
                if(this.q) {
                    link += `&q=${this.q}`;
                }
                if(this.channel_type) {
                    link += `&channel_type=${this.channel_type}`;
                }
                if(this.online) {
                    link += `&online=${this.online}`;
                }
                window.open(link);
            },
            upload() {
                this.$refs["uploadButton"].click();
            },
            onUploadSuccess(res, file, fileList) {
                this.loading = false;
                this.$message({
                    type: "success",
                    message: "上传成功"
                });
                this.$nextTick(() => {
                    this.getChannels();
                });
            },
            onUploadProgress(evt, file, fileList) {
                this.loading = true;
            },
            onUploadError(err, file, fileList) {
                this.loading = false;
                if (err) {
                    this.$message({
                        type: "error",
                        message: err + ""
                    })
                }
            },
            toggleReverse() {
                this.reverse = !this.reverse;
            },
        }, //-- methods
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
