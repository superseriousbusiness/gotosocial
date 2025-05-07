// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package router

import (
	"html/template"
	"testing"
)

func TestOutdentPreformatted(t *testing.T) {
	const html = template.HTML(`
        <div class="text">
            <div
                class="content"
                lang="en"
                title="DW from Arthur is labeled &#34;crawlers&#34;. 
        
            She&#39;s reading a sign on a door that says: &#34;robots.txt: don&#39;t crawl this website, it&#39;s not for you, please, thanks.&#34;
        
        With her hands on her hips looking annoyed she says &#34;That sign won&#39;t stop me because I can&#39;t read!&#34;"
                alt="pee pee poo poo"
            >
                <p>Here's a bunch of HTML, read it and weep, weep then!</p>
                <pre><code class="language-html">&lt;section class=&#34;about-user&#34;&gt;
                    &lt;div class=&#34;col-header&#34;&gt;
                        &lt;h2&gt;About&lt;/h2&gt;
                    &lt;/div&gt;            
                    &lt;div class=&#34;fields&#34;&gt;
                        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;
                        &lt;dl&gt;
                            &lt;div class=&#34;field&#34;&gt;
                                &lt;dt&gt;should you follow me?&lt;/dt&gt;
                                &lt;dd&gt;maybe!&lt;/dd&gt;
                            &lt;/div&gt;
                            &lt;div class=&#34;field&#34;&gt;
                                &lt;dt&gt;age&lt;/dt&gt;
                                &lt;dd&gt;120&lt;/dd&gt;
                            &lt;/div&gt;
                        &lt;/dl&gt;
                    &lt;/div&gt;
                    &lt;div class=&#34;bio&#34;&gt;
                        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;
                        &lt;p&gt;i post about things that concern me&lt;/p&gt;
                    &lt;/div&gt;
                    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;
                        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;
                        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;
                        &lt;span&gt;8 posts.&lt;/span&gt;
                        &lt;span&gt;Followed by 1.&lt;/span&gt;
                        &lt;span&gt;Following 1.&lt;/span&gt;
                    &lt;/div&gt;
                    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;
                        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;
                        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;
                        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
                        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
                    &lt;/div&gt;
                &lt;/section&gt;
                </code></pre>
                <p>There, hope you liked that!</p>
            </div>
        </div>
        <div class="text">
            <div
                class="content"
                lang="en"
                alt="DW from Arthur is labeled &#34;crawlers&#34;. 
        
        She&#39;s reading a sign on a door that says: &#34;robots.txt: don&#39;t crawl this website, it&#39;s not for you, please, thanks.&#34;
        
        With her hands on her hips looking annoyed she says &#34;That sign won&#39;t stop me because I can&#39;t read!&#34;"
            >
                <p>Here's a bunch of HTML, read it and weep, weep then!</p>
                <pre><code class="language-html">&lt;section class=&#34;about-user&#34;&gt;
                    &lt;div class=&#34;col-header&#34;&gt;
                        &lt;h2&gt;About&lt;/h2&gt;
                    &lt;/div&gt;            
                    &lt;div class=&#34;fields&#34;&gt;
                        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;
                        &lt;dl&gt;
                            &lt;div class=&#34;field&#34;&gt;
                                &lt;dt&gt;should you follow me?&lt;/dt&gt;
                                &lt;dd&gt;maybe!&lt;/dd&gt;
                            &lt;/div&gt;
                            &lt;div class=&#34;field&#34;&gt;
                                &lt;dt&gt;age&lt;/dt&gt;
                                &lt;dd&gt;120&lt;/dd&gt;
                            &lt;/div&gt;
                        &lt;/dl&gt;
                    &lt;/div&gt;
                    &lt;div class=&#34;bio&#34;&gt;
                        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;
                        &lt;p&gt;i post about things that concern me&lt;/p&gt;
                    &lt;/div&gt;
                    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;
                        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;
                        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;
                        &lt;span&gt;8 posts.&lt;/span&gt;
                        &lt;span&gt;Followed by 1.&lt;/span&gt;
                        &lt;span&gt;Following 1.&lt;/span&gt;
                    &lt;/div&gt;
                    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;
                        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;
                        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;
                        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
                        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
                    &lt;/div&gt;
                &lt;/section&gt;
                </code></pre>
                <p>There, hope you liked that!</p>
            </div>
        </div>
`)

	const expected = template.HTML(`
        <div class="text">
            <div
                class="content"
                lang="en"
                title="DW from Arthur is labeled &#34;crawlers&#34;. 

    She&#39;s reading a sign on a door that says: &#34;robots.txt: don&#39;t crawl this website, it&#39;s not for you, please, thanks.&#34;

With her hands on her hips looking annoyed she says &#34;That sign won&#39;t stop me because I can&#39;t read!&#34;"
                alt="pee pee poo poo"
            >
                <p>Here's a bunch of HTML, read it and weep, weep then!</p>
<pre><code class="language-html">&lt;section class=&#34;about-user&#34;&gt;
    &lt;div class=&#34;col-header&#34;&gt;
        &lt;h2&gt;About&lt;/h2&gt;
    &lt;/div&gt;            
    &lt;div class=&#34;fields&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;
        &lt;dl&gt;
            &lt;div class=&#34;field&#34;&gt;
&lt;dt&gt;should you follow me?&lt;/dt&gt;
&lt;dd&gt;maybe!&lt;/dd&gt;
            &lt;/div&gt;
            &lt;div class=&#34;field&#34;&gt;
&lt;dt&gt;age&lt;/dt&gt;
&lt;dd&gt;120&lt;/dd&gt;
            &lt;/div&gt;
        &lt;/dl&gt;
    &lt;/div&gt;
    &lt;div class=&#34;bio&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;
        &lt;p&gt;i post about things that concern me&lt;/p&gt;
    &lt;/div&gt;
    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;
        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;
        &lt;span&gt;8 posts.&lt;/span&gt;
        &lt;span&gt;Followed by 1.&lt;/span&gt;
        &lt;span&gt;Following 1.&lt;/span&gt;
    &lt;/div&gt;
    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;
        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;
        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;
        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
    &lt;/div&gt;
&lt;/section&gt;
</code></pre>
                <p>There, hope you liked that!</p>
            </div>
        </div>
        <div class="text">
            <div
                class="content"
                lang="en"
                alt="DW from Arthur is labeled &#34;crawlers&#34;. 

She&#39;s reading a sign on a door that says: &#34;robots.txt: don&#39;t crawl this website, it&#39;s not for you, please, thanks.&#34;

With her hands on her hips looking annoyed she says &#34;That sign won&#39;t stop me because I can&#39;t read!&#34;"
            >
                <p>Here's a bunch of HTML, read it and weep, weep then!</p>
<pre><code class="language-html">&lt;section class=&#34;about-user&#34;&gt;
    &lt;div class=&#34;col-header&#34;&gt;
        &lt;h2&gt;About&lt;/h2&gt;
    &lt;/div&gt;            
    &lt;div class=&#34;fields&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;
        &lt;dl&gt;
            &lt;div class=&#34;field&#34;&gt;
&lt;dt&gt;should you follow me?&lt;/dt&gt;
&lt;dd&gt;maybe!&lt;/dd&gt;
            &lt;/div&gt;
            &lt;div class=&#34;field&#34;&gt;
&lt;dt&gt;age&lt;/dt&gt;
&lt;dd&gt;120&lt;/dd&gt;
            &lt;/div&gt;
        &lt;/dl&gt;
    &lt;/div&gt;
    &lt;div class=&#34;bio&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;
        &lt;p&gt;i post about things that concern me&lt;/p&gt;
    &lt;/div&gt;
    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;
        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;
        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;
        &lt;span&gt;8 posts.&lt;/span&gt;
        &lt;span&gt;Followed by 1.&lt;/span&gt;
        &lt;span&gt;Following 1.&lt;/span&gt;
    &lt;/div&gt;
    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;
        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;
        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;
        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;
    &lt;/div&gt;
&lt;/section&gt;
</code></pre>
                <p>There, hope you liked that!</p>
            </div>
        </div>
`)

	out := outdentPreformatted(html)
	if out != expected {
		t.Fatalf("unexpected output:\n`%s`\n", out)
	}
}
