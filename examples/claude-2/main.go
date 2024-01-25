package main

import (
	"context"
	"fmt"
	"github.com/bincooo/claude-api"
	"github.com/bincooo/claude-api/types"
	"github.com/bincooo/claude-api/vars"
	"time"
)

const (
	attrCtx = `
有一天，一只小狗和一只小猫在公园里玩耍。他们是很好的朋友，经常一起分享食物和玩具。他们也很喜欢探索新的地方，寻找有趣的东西。
这一天，他们发现了一个大箱子，里面装满了各种各样的东西。有帽子、围巾、手套、书本、画笔、玩具车等等。小狗和小猫很好奇，想看看里面还有什么。
他们走进了箱子，开始翻找。他们试戴了帽子，围上了围巾，戴上了手套，觉得很好玩。他们还拿出了书本，看了看上面的字和图画。他们还用画笔在纸上涂涂画画，做出了自己的作品。他们还拿出了玩具车，沿着箱子的边缘开来开去，假装自己是司机。
他们玩得很开心，忘记了时间。直到天色渐暗，他们才想起要回家。他们收拾好自己的东西，准备离开。
就在这时，他们听到了一阵响声。原来是箱子的主人回来了。他是一个老爷爷，住在公园附近的一所小屋里。他每天都会来公园散步，顺便把自己不用的东西放在箱子里，希望能给别人带来一些快乐。
老爷爷看到了箱子里的小狗和小猫，感到很惊讶。他问他们：“你们是谁？你们为什么在我的箱子里？”
小狗和小猫吓得不敢说话，只是低头看着自己的脚。老爷爷看出了他们的害怕，笑着说：“不要害怕，我不会伤害你们。你们喜欢我的东西吗？”
小狗和小猫点点头，说：“是的，我们很喜欢。谢谢您。”
老爷爷说：“不用谢。我很高兴你们能喜欢我的东西。你们可以随便玩，只要不弄坏就行。你们想留下来吃饭吗？我有很多好吃的东西。”
小狗和小猫听了，眼睛都亮了起来。他们说：“真的吗？那太好了！我们很饿。”
老爷爷说：“那就跟我走吧。”说完，他把箱子关上，牵着小狗和小猫走向自己的小屋。
从那以后，小狗和小猫就成了老爷爷的新朋友。他们每天都会来公园玩耍，并且和老爷爷一起分享食物和故事。他们过得很快乐，也让老爷爷感到很温暖。
`
)

func main() {
	var (
		token = "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"
	)
	// email, tk, err := util.LoginFor("", "gmail.com", "http://127.0.0.1:7890")
	//if err != nil {
	//	panic(err)
	//}
	//token = tk
	//fmt.Println(email)
	options := claude.NewDefaultOptions(token, "", vars.Model4WebClaude2)
	//options.Agency = "http://127.0.0.1:7890"
	options.BaseURL = "https://bincooo-single-proxy.hf.space/api"
	chat, err := claude.New(options)
	if err != nil {
		panic(err)
	}

	prompt := "hi"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "who are you?"
	fmt.Println("You: ", prompt)
	partialResponse, err = chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "总结附件内容："
	fmt.Println("You: ", prompt)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	partialResponse, err = chat.Reply(ctx, prompt, []types.Attachment{
		{
			Content:  attrCtx,
			FileName: "paste.txt",
			FileSize: 999999,
			FileType: "txt",
		},
	})
	if err != nil {
		panic(err)
	}
	Println(partialResponse)
}

func Println(partialResponse chan types.PartialResponse) {
	for {
		message, ok := <-partialResponse
		if !ok {
			return
		}

		if message.Error != nil {
			panic(message.Error)
		}

		fmt.Println(message.Text)
		fmt.Println("===============")
	}
}
