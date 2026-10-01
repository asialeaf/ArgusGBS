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
                            <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '管理员')">
                                <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                                <div class="input-group-btn">
                                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出级联通道" :disabled="shareAllChannel"><i class="fa fa-download"></i></button>
                                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="upload" title="导入级联通道" :disabled="shareAllChannel"><i class="fa fa-upload"></i></button>
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
                                <option value="dir">子目录</option>
                            </select>
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm" v-if="channel_type == 'device'">
                            <label>在线状态</label>
                            <select class="form-control" v-model.trim="online">
                                <option value="">全部</option>
                                <option value="true">在线</option>
                                <option value="false">离线</option>
                            </select>
                        </div>
                        <span class="hidden-xs" v-if="channel_type == 'device'">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm">
                            <div class="checkbox" style="user-select:none;">
                                <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model.trim="related" name="Related" :disabled="shareAllChannel">
                                    只看{{reverse ? '未选' : '已选'}}({{relateCnt}})
                                </el-checkbox>
                                <span :style="shareAllChannel ? 'cursor:not-allowed;' : ''">
                                    <a role="button" @click="toggleReverse" :style="{'pointer-events': shareAllChannel ? 'none' : 'all'}" :class="{'text-gray': shareAllChannel}">
                                        <i class="fa fa-sort"></i>
                                    </a>
                                </span>
                            </div>
                        </div>
                        <span class="hidden-xs">&nbsp;&nbsp;</span>
                        <div class="form-group form-group-sm" v-if="!userInfo || userInfo.HasAllChannel">
                            <div class="checkbox">
                                <el-checkbox style="margin-top:-5px;padding-left:0;" size="small" v-model="shareAllChannel" @change="toggleShareAllChannel" name="ShareAllChannel">全部共享</el-checkbox>
                            </div>
                        </div>
                        <div class="form-group form-group-sm pull-right">
                            <div class="btn-group btn-group-sm">
                                <button type="button" class="btn btn-danger" @click.prevent="removeAll" v-if="(!userInfo || userInfo.HasAllChannel) && !q && !channel_type && !reverse && !shareAllChannel && relateCnt"><i class="fa fa-remove"></i> 清空</button>
                                <button type="button" class="btn btn-primary" @click.prevent="checkRepeat"><i class="fa fa-check"></i> 查重</button>
                            </div>
                        </div>
                    </div>
                    <br>
                    <el-table :data="channels" stripe @sort-change="sortChange" @select="select" @select-all="selectAll" :max-height="500"
                        ref="channelTable" v-loading="loading" element-loading-text="加载中...">
                        <el-table-column type="selection" width="55" fixed :selectable="selectable"></el-table-column>
                        <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="ID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                        <el-table-column prop="CustomID" label="自定义通道国标编号" min-width="200" show-overflow-tooltip sortable="custom">
                            <template slot-scope="props">
                                <a role="button" :class="{'text-orange': !!props.row.CustomID}" @click.prevent="setChannelID(props.row, true, $event)" v-if="!props.row.Editing">{{props.row.CustomID || props.row.ID}}</a>
                                <input type="text" style="width:170px;padding:2px;line-height:100%;" oninput="value=value.replace(/[^\d]/g,'')"
                                    @keydown.esc.stop.prevent="setChannelID(props.row, false, $event)"
                                    @keydown.enter.stop.prevent="setChannelID(props.row, false, $event)"
                                    @blur="setChannelID(props.row, false, $event)" :value="props.row.CustomID || props.row.ID"
                                    v-else v-focus>
                            </template>
                        </el-table-column>
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
        <el-upload :action="`/api/v1/cascade/channel/import?id=${id}`" accept=".xlsx" class="hide"
            :show-file-list="false"
            :on-success="onUploadSuccess" :on-error="onUploadError" :on-progress="onUploadProgress">
            <a role="button" title="上传级联通道" ref="uploadButton">
                <i class="fa fa-upload"></i> 上传级联通道
            </a>
        </el-upload>
        <el-dialog :title="`重复通道(${name||id})`" width="50%" center append-to-body :lock-scroll="false" :visible.sync="repeatVisible" :before-close="handleRepeatClose">
            <el-table :data="repeatChannels" stripe @sort-change="repeatSortChange"  ref="repeatTable" v-loading="repeatLoading" element-loading-text="加载中..." empty-text="暂无重复通道">
                <el-table-column prop="RepeatCode" label="通道编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                <el-table-column label="操作" min-width="180" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                    <template slot-scope="props">
                        <div class="btn-group btn-group-xs">
                            <button type="button" class="btn btn-default" v-clipboard="props.row.RepeatCode" @success="$message({type:'success', message:'成功拷贝到粘贴板'})">
                                <i class="fa fa-copy"></i> 拷贝编号
                            </button>
                            <button type="button" class="btn btn-info" @click.prevent="queryRepeat(props.row.RepeatCode)">
                                <i class="fa fa-info"></i> 查看详情
                            </button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="RepeatCount" label="关联次数" min-width="160" show-overflow-tooltip sortable="custom"></el-table-column>
            </el-table>
            <el-pagination v-if="repeatTotal > 0" layout="total,prev,pager,next,sizes" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="repeatTotal" :page-size.sync="repeatPageSize" :current-page.sync="repeatCurrentPage"></el-pagination>
            <div class="clearfix"></div>
        </el-dialog>
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
                shareAllChannel: false,
                loading: false,
                channels: [],
                selection: [],
                id: '', // 外部关联 id
                name: '', // 外部关联 name
                repeatVisible: false,
                repeatChannels: [],
                repeatTotal: 0,
                repeatPageSize: 10,
                repeatCurrentPage: 1,
                repeatLoading: false,
                repeatSort: "",
                repeatOrder: "",
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
            },
            repeatCurrentPage: function(newVal, oldVal) {
                this.doRepeatSearch(newVal);
            },
            repeatPageSize: function(newVal, oldVal) {
                this.doRepeatSearch();
            },
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
        directives: {
            focus: {
                inserted: function (el) {
                    el.focus();
                    el.select();
                }
            }
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
                    $.post("/api/v1/cascade/savechannels", {
                        id: this.id,
                        channels: [`${row.DeviceID}:${row.ID}`],
                    }).always(() => {
                        this.getChannels();
                    })
                } else {
                    $.post("/api/v1/cascade/removechannels", {
                        id: this.id,
                        channels: [`${row.DeviceID}:${row.ID}`],
                    }).always(() => {
                        this.getChannels();
                    })
                }
            },
            removeAll() {
                $(this.$el).hide();
                this.$confirm(`确认清空已逐条关联的通道?`, "提示", {
                    lockScroll: false,
                }).then(() => {
                    $(this.$el).show().focus();
                    $.post("/api/v1/cascade/removechannels", {
                        id: this.id,
                        channels: ["*"],
                    }).always(() => {
                        this.getChannels();
                    })
                }).catch(() => {
                    $(this.$el).show().focus();
                });
            },
            selectAll(selection) {
                if(this.shareAllChannel) return;
                var keys = [];
                if(selection.length) {
                    for(var row of selection) {
                        var idx = this.selection.indexOf(row);
                        if(idx < 0) {
                            keys.push(`${row.DeviceID}:${row.ID}`);
                        }
                    }
                    $.post("/api/v1/cascade/savechannels", {
                        id: this.id,
                        channels: keys,
                    }).always(() => {
                        this.getChannels();
                    })
                } else {
                    for(var row of this.selection) {
                        keys.push(`${row.DeviceID}:${row.ID}`);
                    }
                    $.post("/api/v1/cascade/removechannels", {
                        id: this.id,
                        channels: keys,
                    }).always(() => {
                        this.getChannels();
                    })
                }
            },
            doSearch(page = 1) {
                this.currentPage = page;
                if(!this.repeatVisible) {
                    this.getChannels();
                }
            },
            doDelaySearch: _.debounce(function() {
                this.doSearch();
            }, 800),
            doRepeatSearch(page = 1) {
                this.repeatCurrentPage = page;
                if(this.repeatVisible) {
                    this.loadRepeat();
                }
            },
            doRepeatDelaySearch: _.debounce(function() {
                this.doRepeatSearch();
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
                if(!this.shareAllChannel) {
                    return true;
                }
                return false;
            },
            getChannels() {
                if(!this.id) return;
                this.loading = true;
                $.get("/api/v1/cascade/channellist", {
                    id: this.id,
                    q: this.q,
                    start: (this.currentPage -1) * this.pageSize,
                    limit: this.pageSize,
                    channel_type: this.channel_type,
                    online: this.channel_type == 'device' ? this.online : '',
                    related: this.related,
                    reverse: this.reverse,
                    sort: this.sort,
                    order: this.order,
                }).then(ret => {
                    this.$refs["channelTable"].clearSelection();
                    this.total = ret.ChannelCount;
                    this.relateCnt = ret.ChannelRelateCount;
                    this.shareAllChannel = !!ret.ShareAllChannel;
                    this.channels = ret.ChannelList || [];
                    this.selection = [];
                    this.$nextTick(() => {
                        this.channels.forEach(row => {
                            var sel = row.CascadeID != "";
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
                this.name = '';
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
                this.shareAllChannel = false;
                this.currentPage = 1;
                this.pageSize = 10;
            },
            toggleShareAllChannel(val) {
                $.post("/api/v1/cascade/setshareallchannel", {
                    id: this.id,
                    shareallchannel: val,
                }).always(() => {
                    this.doSearch();
                })
            },
            setChannelID(row, editing, event) {
                var _editing = row.Editing;
                this.$set(row, "Editing", editing);
                if(_editing && !editing) {
                    var val = event.target.value.trim();
                    if (val == row.ID) {
                        val = "";
                    }
                    if(row.CustomID != val) {
                        $.post('/api/v1/device/setchannelid', {
                            serial: row.DeviceID,
                            code: row.ID,
                            id: val,
                        }).then(data => {
                            row.CustomID = val;
                        });
                    }
                }
            },
            show(id, name='') {
                this.id = id;
                this.name = name;
                $(this.$el).modal("show");
                this.getChannels();
            },
            hide() {
                $(this.$el).modal("hide");
            },
            download() {
                let link = `/api/v1/cascade/channel/export?id=${this.id}`;
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
            checkRepeat() {
                $(this.$el).hide();
                this.repeatVisible = true;
                this.loadRepeat();
            },
            loadRepeat() {
                this.repeatLoading = true;
                $.get("/api/v1/cascade/channel/repeat", {
                    id: this.id,
                    start: (this.repeatCurrentPage -1) * this.repeatPageSize,
                    limit: this.repeatPageSize,
                    sort: this.repeatSort,
                    order: this.repeatOrder,
                }).then(ret => {
                    this.repeatChannels = ret.RepeatList || [];
                    this.repeatTotal = ret.RepeatTotal;
                }).always(() => {
                    this.repeatLoading = false;
                })
            },
            queryRepeat(code) {
                this.channel_type = "";
                this.online = "";
                this.reverse = false;
                if(!this.shareAllChannel) {
                    this.related = true;
                }
                this.repeatVisible = false;
                $(this.$el).show().focus();
                this.q = code;
            },
            repeatSortChange(data) {
                this.repeatSort = data.prop;
                this.repeatOrder = data.order == "ascending" ? "asc" : "desc";
                this.loadRepeat();
            },
            handleRepeatClose(done) {
                done();
                this.repeatTotal = 0;
                this.repeatChannels = [];
                this.repeatCurrentPage = 1;
                this.repeatPageSize = 10;
                $(this.$el).show().focus();
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
