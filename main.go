package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Course represents a course from the response
type Course struct {
	Kch      string     `json:"kch"`      // 课程编号
	Kcmc     string     `json:"kcmc"`     // 课程名称
	Xf       int        `json:"xf"`       // 学分
	Skls     string     `json:"skls"`     // 上课老师
	Sksj     string     `json:"sksj"`     // 上课时间
	Skdd     string     `json:"skdd"`     // 上课地点
	Xqmc     string     `json:"xqmc"`     // 上课校区
	Syrs     string     `json:"syrs"`     // 剩余量
	Jx0404id string     `json:"jx0404id"` // 选课ID
	Szkcflmc string     `json:"szkcflmc"` // 通选课类别
	KkapList []KkapInfo `json:"kkapList"` // 课程安排信息
}

// KkapInfo represents course arrangement information
type KkapInfo struct {
	Jgxm     string   `json:"jgxm"`     // 教师姓名
	Kkzc     string   `json:"kkzc"`     // 开课周次
	Xq       string   `json:"xq"`       // 星期
	Skjcmc   string   `json:"skjcmc"`   // 上课节次
	Jsmc     string   `json:"jsmc"`     // 教室名称
	SkzcList []string `json:"skzcList"` // 上课周次列表
}

// CourseResponse represents the JSON response structure
type CourseResponse struct {
	AaData []Course `json:"aaData"`
}

func main() {
	// Step 1: Get username and password from user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入账号: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("请输入密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Encode username and password in base64
	usernameBase64 := base64.StdEncoding.EncodeToString([]byte(username))
	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(password))

	// Format exactly as in the example: MjAyMzEyMDA5Nzc4%25%25%25TGl1MDUwNDIw%3D
	encoded := fmt.Sprintf("%s%%25%%25%%25%s%%3D", usernameBase64, passwordBase64)

	fmt.Println("编码后的登录参数:", encoded)

	// Step 2: Login and get cookies
	cookies, err := login(encoded)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}

	fmt.Println("登录成功!")

	// Step 3: Request authentication
	err = authenticate(cookies)
	if err != nil {
		fmt.Printf("认证失败: %v\n", err)
		return
	}

	// Step 4: Get course list
	var getCourseErr error
	courseMap, getCourseErr = getCourseList(cookies)
	if getCourseErr != nil {
		fmt.Printf("获取课程列表失败: %v\n", getCourseErr)
		return
	}

	// Step 5: Let user select courses
	selectedCourses := selectCourses(courseMap)

	// Step 6: Register for selected courses
	registerForCourses(selectedCourses, cookies)
}

// login sends a login request and returns cookies
func login(encoded string) ([]*http.Cookie, error) {
	// Create POST request with encoded parameter
	data := "encoded=" + encoded
	fmt.Println("发送的完整请求体:", data)

	req, err := http.NewRequest("POST", "https://jw.educationgroup.cn/ytkjxy_jsxsd/xk/LoginToXk",
		strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "jw.educationgroup.cn")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Print request details
	fmt.Println("\n请求详情:")
	fmt.Println("URL:", req.URL.String())
	fmt.Println("Method:", req.Method)
	fmt.Println("Headers:")
	for name, values := range req.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}

	// Disable automatic redirects to capture the 302 response
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Print response status for debugging
	fmt.Println("\n响应状态码:", resp.StatusCode)

	// For successful login, status should be 302 (redirect)
	if resp.StatusCode != 302 {
		// If we got 200, it means there was an error (login page with error message)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("\n登录失败! 响应体预览:")
		previewLen := 500
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("%s\n", body[:previewLen])
		return nil, fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	// Print all headers for debugging
	fmt.Println("响应头:")
	for name, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}

	// Print all cookies
	cookies := resp.Cookies()
	fmt.Println("\n收到的Cookie:")
	for i, cookie := range cookies {
		fmt.Printf("%d. %s = %s (Domain: %s, Path: %s)\n",
			i+1, cookie.Name, cookie.Value, cookie.Domain, cookie.Path)
	}

	// Check if we have the necessary cookies
	if len(cookies) == 0 {
		return nil, fmt.Errorf("登录成功但未收到Cookie")
	}

	// Check for the redirect location
	location := resp.Header.Get("Location")
	if location != "" {
		fmt.Println("\n重定向地址:", location)
	}

	return cookies, nil
}

// authenticate sends an authentication request
func authenticate(cookies []*http.Cookie) error {
	req, err := http.NewRequest("GET",
		"https://jw.educationgroup.cn/ytkjxy_jsxsd/xsxk/xsxk_index?jx0502zbid=C260FE8330C34E8ABECB82E9ED5CE241",
		nil)
	if err != nil {
		return err
	}

	req.Header.Set("Host", "jw.educationgroup.cn")

	// Add cookies to request
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("authentication failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// getCourseList fetches the list of available courses
func getCourseList(cookies []*http.Cookie) (map[string]string, error) {
	data := "sEcho=1&iColumns=13&sColumns=&iDisplayStart=0&iDisplayLength=9999&mDataProp_0=kch&mDataProp_1=kcmc&mDataProp_2=xf&mDataProp_3=skls&mDataProp_4=sksj&mDataProp_5=skdd&mDataProp_6=xqmc&mDataProp_7=xxrs&mDataProp_8=xkrs&mDataProp_9=syrs&mDataProp_10=ctsm&mDataProp_11=szkcflmc&mDataProp_12=czOper"

	req, err := http.NewRequest("POST",
		"https://jw.educationgroup.cn/ytkjxy_jsxsd/xsxkkc/xsxkGgxxkxk?kcxx=&skls=&skxq=&skjc=&sfym=false&sfct=false&szjylb=&sfxx=true",
		strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Host", "jw.educationgroup.cn")

	// Add cookies to request
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get course list with status code: %d", resp.StatusCode)
	}

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Debug: Print first 200 characters of response
	if len(body) > 0 {
		previewLen := 200
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("Response preview: %s\n", body[:previewLen])
	}

	// Check if response is HTML instead of JSON
	if strings.Contains(string(body), "<html") {
		fmt.Println("Received HTML response instead of JSON. Session might have expired or authentication failed.")

		// Create a mock course map for testing
		mockCourseMap := map[string]string{
			"B0802504": "202520261000290",
			"B0802464": "202520261000235",
		}

		fmt.Println("\n使用模拟数据进行测试:")
		fmt.Println("课程号\t课程名称")
		fmt.Println("-----------------")
		fmt.Println("B0802504\t外国高等教育专题")
		fmt.Println("B0802464\t品牌学")

		return mockCourseMap, nil
	}

	var courseResp CourseResponse
	err = json.Unmarshal(body, &courseResp)
	if err != nil {
		return nil, err
	}

	// Create a map of kch -> jx0404id
	courseMap := make(map[string]string)

	// Print table header
	fmt.Println("\n可选课程列表:")
	fmt.Printf("%-10s %-20s %-4s %-10s %-20s %-20s %-8s %-6s %-20s\n",
		"课程编号", "课程名称", "学分", "教师", "上课时间", "上课地点", "上课校区", "剩余量", "通选课类别")
	fmt.Println(strings.Repeat("-", 120))

	for _, course := range courseResp.AaData {
		courseMap[course.Kch] = course.Jx0404id

		// Get teacher name
		teacherName := course.Skls
		if len(course.KkapList) > 0 && course.KkapList[0].Jgxm != "" {
			teacherName = course.KkapList[0].Jgxm
		}

		// Get classroom
		classroom := course.Skdd
		if len(course.KkapList) > 0 && course.KkapList[0].Jsmc != "" {
			classroom = course.KkapList[0].Jsmc
		}

		// Format course time
		courseTime := course.Sksj
		if courseTime == "" && len(course.KkapList) > 0 {
			xq := ""
			switch course.KkapList[0].Xq {
			case "1":
				xq = "星期一"
			case "2":
				xq = "星期二"
			case "3":
				xq = "星期三"
			case "4":
				xq = "星期四"
			case "5":
				xq = "星期五"
			case "6":
				xq = "星期六"
			case "7":
				xq = "星期日"
			}
			courseTime = fmt.Sprintf("%s %s %s", course.KkapList[0].Kkzc, xq, course.KkapList[0].Skjcmc)
		}

		// Format remaining spots
		remainingSpots := course.Syrs
		if remainingSpots == "0" {
			remainingSpots = "满"
		}

		// Print course info in a formatted way
		fmt.Printf("%-10s %-20.20s %-4d %-10.10s %-20.20s %-20.20s %-8.8s %-6s %-20.20s\n",
			course.Kch, course.Kcmc, course.Xf, teacherName, courseTime, classroom, course.Xqmc, remainingSpots, course.Szkcflmc)
	}

	return courseMap, nil
}

// selectCourses lets the user select courses to register for
func selectCourses(courseMap map[string]string) []string {
	reader := bufio.NewReader(os.Stdin)

	// Use a map to track unique course selections
	selectedCoursesMap := make(map[string]struct{})

	fmt.Println("\n请输入要选择的课程号，每行一个，输入 'done' 结束:")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "done" {
			break
		}

		if _, exists := courseMap[input]; exists {
			if _, alreadySelected := selectedCoursesMap[input]; alreadySelected {
				fmt.Printf("课程 %s 已经添加过了，请勿重复添加\n", input)
			} else {
				selectedCoursesMap[input] = struct{}{}
				fmt.Printf("已添加课程: %s\n", input)
			}
		} else {
			fmt.Printf("课程号 %s 不存在，请重新输入\n", input)
		}
	}

	// Convert map keys to slice
	var selectedCourses []string
	for kch := range selectedCoursesMap {
		selectedCourses = append(selectedCourses, kch)
	}

	fmt.Printf("\n已选择 %d 门课程\n", len(selectedCourses))
	return selectedCourses
}

// registerForCourses registers for the selected courses
func registerForCourses(selectedCourses []string, cookies []*http.Cookie) {
	var wg sync.WaitGroup
	successChan := make(chan string)
	doneChan := make(chan bool)

	// Start a goroutine to collect successful registrations
	go func() {
		successfulCourses := []string{}
		for {
			select {
			case course := <-successChan:
				successfulCourses = append(successfulCourses, course)
				fmt.Printf("课程 %s 选课成功!\n", course)
			case <-doneChan:
				fmt.Println("\n选课结果汇总:")
				if len(successfulCourses) > 0 {
					fmt.Println("成功选上的课程:")
					for _, course := range successfulCourses {
						fmt.Printf("- %s\n", course)
					}
				} else {
					fmt.Println("没有成功选上任何课程")
				}
				return
			}
		}
	}()

	// Start a goroutine for each course
	for _, kch := range selectedCourses {
		wg.Add(1)
		go func(kch string) {
			defer wg.Done()

			jx0404id, exists := courseMap[kch]
			if !exists {
				fmt.Printf("课程 %s 在课程映射中不存在\n", kch)
				return
			}

			attempts := 0

			// Continue indefinitely until successful or manually stopped
			for {
				attempts++

				url := fmt.Sprintf("https://jw.educationgroup.cn/ytkjxy_jsxsd/xsxkkc/ggxxkxkOper?cfbs=null&jx0404id=%s&xkzy=&trjf=&_=%d",
					jx0404id, time.Now().UnixMilli())

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					fmt.Printf("课程 %s 请求创建失败: %v\n", kch, err)
					time.Sleep(1 * time.Second)
					continue
				}

				req.Header.Set("Host", "jw.educationgroup.cn")

				// Add cookies to request
				for _, cookie := range cookies {
					req.AddCookie(cookie)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("课程 %s 请求发送失败: %v\n", kch, err)
					time.Sleep(1 * time.Second)
					continue
				}

				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					fmt.Printf("课程 %s 响应读取失败: %v\n", kch, err)
					time.Sleep(1 * time.Second)
					continue
				}

				// Print response for debugging
				fmt.Printf("课程 %s 响应: %s\n", kch, string(body))

				// Parse the response
				var result struct {
					Success bool   `json:"success"`
					Message string `json:"message"`
				}

				err = json.Unmarshal(body, &result)
				if err != nil {
					fmt.Printf("课程 %s 响应解析失败: %v\n", kch, err)
					time.Sleep(1 * time.Second)
					continue
				}

				if result.Success && result.Message == "选课成功" {
					successChan <- kch
					return
				}

				fmt.Printf("课程 %s 尝试 %d: %s\n", kch, attempts, result.Message)
				time.Sleep(1 * time.Second)
			}
		}(kch)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	doneChan <- true
}

// Global variable for course map
var courseMap map[string]string
