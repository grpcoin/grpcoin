<!doctype html>
<html lang="en">

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.1/dist/css/bootstrap.min.css" rel="stylesheet"
        integrity="sha384-+0n0xVW2eSR5OomGNYDnhzAbDsOXxcvSN1TPprVMTNDbiYZCxYbOOl7+AMvyTG2x" crossorigin="anonymous">
    <title>{{.}}</title>
    <style>
        a {
            text-decoration: none;
            color: #f1f2f6;
        }

        a:hover {
            color: #ced6e0;
        }

        body {
            background-color: #1A202C;
        }

        .bg-color-black {
            background-color: #2C313D !important;
        }

        .table-bg {
            background-color: #373C47;
        }

        table>tbody>tr:nth-of-type(odd) {
            background-color: #2C313D !important;
            color: whitesmoke !important;
        }

        table>tbody>tr:nth-of-type(even) {
            background-color: #373C47 !important;
            color: whitesmoke !important;
        }

        .bg-hover:hover {
            background-color: #373C47 !important;
            color: whitesmoke !important;
        }

        /* width */
        ::-webkit-scrollbar {
            width: 10px;
        }

        ::-webkit-scrollbar-corner {
            background: rgba(0, 0, 0, 0);
        }

        /* Handle */
        ::-webkit-scrollbar-thumb {
            background: #6A6B6D;
            border-radius: 10px;
        }

        .gh-bg-color-red {
            background-color: #b33939 !important;
        }

        .gh-bg-color-green {
            background-color: #218c74 !important;
        }

        .returns-tbl:nth-of-type(even) {
            border-right: #f1f2f6;
            border-right-style: inset;
            border-left: #f1f2f6;
            border-left-style: inset;
        }

        .returns-tbl {
            border-bottom: none;
        }

        .returns-tbl:last-child {
            border-right-style: none;
        }

        .leaderboard tbody {
            counter-reset: rowNumber;
        }

        .leaderboard tr {
            counter-increment: rowNumber;
        }

        .leaderboard tr td:first-child::before {
            content: counter(rowNumber);
            min-width: 1em;
            margin-right: 0.5em;
        }

        #chart .btn:focus {
            /* chart time buttons do not focus off in macOS */
            outline: none;
            box-shadow: none;
        }
    </style>
</head>

<body>
