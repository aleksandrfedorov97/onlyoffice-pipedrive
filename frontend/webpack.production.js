/* eslint-disable */
const { merge } = require("webpack-merge");
const webpack = require('webpack');
const dotenv = require('dotenv');
const common = require("./webpack.common.js");

module.exports = merge(common, {
    mode: "production",
    plugins: [
        new webpack.DefinePlugin({
            'process.env.PIPEDRIVE_CREATE_MODAL_ID': JSON.stringify(process.env.PIPEDRIVE_CREATE_MODAL_ID),
        }),
    ],
});
