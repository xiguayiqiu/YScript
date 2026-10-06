<?php
$result = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $name = isset($_POST['name']) ? trim($_POST['name']) : '';
    $message = isset($_POST['message']) ? trim($_POST['message']) : '';

    if ($name === '' || $message === '') {
        $result = '<p class="error">姓名和消息不能为空。</p>';
    } else {
        $safeName = htmlspecialchars($name, ENT_QUOTES, 'UTF-8');
        $safeMessage = htmlspecialchars($message, ENT_QUOTES, 'UTF-8');
        $result = '<p class="success">提交成功！</p>'
                . '<p><strong>姓名：</strong>' . $safeName . '</p>'
                . '<p><strong>消息：</strong>' . $safeMessage . '</p>';
    }
}
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>前端三剑客 + PHP 测试页</title>
    <style>
        body {
            margin: 0;
            font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
            background: #f4f6fb;
            display: flex;
            justify-content: center;
            padding: 40px 16px;
        }
        .card {
            background: #fff;
            width: 100%;
            max-width: 460px;
            padding: 24px;
            border-radius: 12px;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
        }
        h1 {
            font-size: 20px;
            margin: 0 0 16px;
        }
        label {
            display: block;
            margin-bottom: 6px;
            font-size: 14px;
            color: #555;
        }
        input, textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 8px;
            font-size: 14px;
            margin-bottom: 14px;
        }
        textarea {
            height: 100px;
            resize: vertical;
        }
        button {
            padding: 10px 18px;
            border: none;
            border-radius: 8px;
            background: #4f46e5;
            color: #fff;
            font-size: 14px;
            cursor: pointer;
        }
        button:hover {
            background: #4338ca;
        }
        .result-box {
            margin-top: 18px;
            padding: 14px;
            border-radius: 8px;
            background: #f9fafb;
            display: none;
        }
        .error {
            color: #b91c1c;
        }
        .success {
            color: #166534;
        }
    </style>
</head>
<body>
    <div class="card">
        <h1>前端三剑客 + PHP 测试</h1>

        <form id="testForm" method="POST" action="test.php">
            <label for="name">姓名</label>
            <input type="text" id="name" name="name" placeholder="请输入姓名">

            <label for="message">消息</label>
            <textarea id="message" name="message" placeholder="请输入消息内容"></textarea>

            <button type="submit">提交</button>
        </form>

        <div class="result-box" id="resultBox">
            <?php echo $result; ?>
        </div>
    </div>

    <script>
        const form = document.getElementById('testForm');
        const resultBox = document.getElementById('resultBox');

        form.addEventListener('submit', function (e) {
            const name = document.getElementById('name').value.trim();
            const message = document.getElementById('message').value.trim();

            if (name === '' || message === '') {
                e.preventDefault();
                alert('姓名和消息不能为空');
                return;
            }

            resultBox.style.display = 'block';
        });

        // 如果 PHP 已经返回了结果，就显示结果区域
        if (resultBox.innerHTML.trim() !== '') {
            resultBox.style.display = 'block';
        }
    </script>
</body>
</html>
