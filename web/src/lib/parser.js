/**
 * src/lib/parser.js
 * 
 * Takes a web-tree-sitter `Tree` and maps it to the generic AST-JSON
 * format expected by the Go-WASM layout engine.
 */

const MAXLEN = 64;
function oneline(t) {
    if (!t) return "?";
    t = t.replace(/\s+/g, " ").trim();
    return t.length <= MAXLEN ? t : t.substring(0, MAXLEN - 1) + "…";
}

function collectDefinedFunctions(node, defined) {
    if (!node) return;
    if (node.type === "function_definition") {
        let nameNode = node.childForFieldName("name");
        if (!nameNode && node.childForFieldName("declarator")) {
             let decl = node.childForFieldName("declarator");
             while (decl && decl.type !== "identifier" && decl.type !== "function_declarator") {
                 decl = decl.childForFieldName("declarator");
             }
             if (decl && decl.type === "function_declarator") {
                 let n = decl.childForFieldName("declarator");
                 if (n && n.type === "identifier") nameNode = n;
             }
        }
        if (nameNode) defined.add(nameNode.text);
    }
    for (let i = 0; i < node.childCount; i++) {
        collectDefinedFunctions(node.child(i), defined);
    }
}

function hasInputNode(node) {
    if (!node) return false;
    if (node.type === "call" || node.type === "call_expression") {
        const func = node.childForFieldName("function");
        if (func && (func.text === "input" || func.text === "cin" || func.text === "std::cin")) return true;
    }
    if (node.text === "cin" || node.text === "std::cin") return true;
    for (let i = 0; i < node.namedChildCount; i++) {
        if (hasInputNode(node.namedChildren[i])) return true;
    }
    return false;
}

function getCallName(node) {
    if (!node) return null;
    if (node.type === "call" || node.type === "call_expression") {
        const func = node.childForFieldName("function");
        if (func) return func.text;
    }
    return null;
}

// endof — верхня межа range як «stop - 1», зі згортанням «X + 1» → «X» і числових
// констант (range(n) → 0, n-1). Дзеркалить endof() з parser.py (щоб веб і CLI збігались).
function endof(node) {
    if (!node) return "?";
    if (node.type === "integer") {
        const n = parseInt(node.text, 10);
        if (!Number.isNaN(n)) return String(n - 1);
    }
    if (node.type === "binary_operator") {
        const op = node.childForFieldName("operator");
        const right = node.childForFieldName("right");
        const left = node.childForFieldName("left");
        if (op && op.text === "+" && right && right.text === "1" && left) return left.text;
    }
    return node.text + " - 1";
}

// isTrueNode/breakIfNode — для розпізнавання «while True» (нескінченний цикл) та
// ідіоми післяумови «while True: … if C: break» (do-while). Дзеркалить is_true/break_if з parser.py.
function isTrueNode(n) {
    return !!n && (n.type === "true" || (n.type === "integer" && n.text === "1"));
}
function breakIfNode(st) {
    if (!st || st.type !== "if_statement" || st.childForFieldName("alternative")) return null;
    const body = st.childForFieldName("consequence");
    if (!body) return null;
    const kids = body.namedChildren.filter(c => c.type !== "comment");
    if (kids.length === 1 && kids[0].type === "break_statement") return st.childForFieldName("condition");
    return null;
}

function formatArg(argNode) {
    if (argNode.type === "string") {
         let text = argNode.text;
         if (text.startsWith('f"') || text.startsWith("f'")) {
             return `«${text}»`;
         }
         if (text.length >= 2 && (text[0] === '"' || text[0] === "'")) {
             return `«${text.substring(1, text.length - 1)}»`;
         }
    }
    return argNode.text;
}

export function parseTreeToAstJson(tree, lang) {
    const defined = new Set();
    collectDefinedFunctions(tree.rootNode, defined);
    
    const isCpp = lang === "cpp";
    const isPy = lang === "python";

    function stmt(s) {
        if (!s) return null;

        if (s.type === "if_statement" || s.type === "elif_clause") {
            const cond = s.childForFieldName("condition");
            const conseq = s.childForFieldName("consequence") || s.childForFieldName("body");
            let alt = s.childForFieldName("alternative");
            
            if (alt && alt.type === "else_clause") {
                let altBody = alt.childForFieldName("body") || alt.childForFieldName("consequence");
                if (!altBody && alt.namedChildCount > 0) {
                     altBody = alt.namedChildren[0];
                }
                alt = altBody;
            }
            
            let condText = cond ? cond.text : "?";
            if (isCpp && condText.startsWith("(") && condText.endsWith(")")) {
                condText = condText.substring(1, condText.length - 1);
            }
            
            return {
                kind: "if",
                cond: oneline(condText),
                then: block(conseq),
                else: block(alt)
            };
        }
        
        if (s.type === "while_statement") {
            const cond = s.childForFieldName("condition");
            const body = s.childForFieldName("body");

            // while True: … — нескінченний цикл (без ромба); якщо останнє «if C: break»
            // — це цикл з післяумовою (do-while). Як у parser.py.
            if (isPy && isTrueNode(cond) && body) {
                const stmts = body.namedChildren.filter(c => c.type !== "comment");
                const last = stmts[stmts.length - 1];
                const brkCond = breakIfNode(last);
                if (brkCond) {
                    return { kind: "dowhile", cond: oneline(brkCond.text),
                        body: { kind: "block", stmts: stmts.slice(0, -1).map(stmt).filter(Boolean) } };
                }
                return { kind: "infloop", body: block(body) };
            }

            let condText = cond ? cond.text : "?";
            if (isCpp && condText.startsWith("(") && condText.endsWith(")")) {
                condText = condText.substring(1, condText.length - 1);
            }
            
            return {
                kind: "while",
                cond: oneline(condText),
                body: block(body),
                else: block(s.childForFieldName("alternative"))
            };
        }
        
        if (s.type === "for_statement") {
            if (isPy) {
                const left = s.childForFieldName("left");
                const right = s.childForFieldName("right");
                let condText = `${left ? left.text : "?"} ∈ ${right ? right.text : "?"}`;
                
                if (right && (right.type === "call" || right.type === "call_expression")) {
                    const func = getCallName(right);
                    if (func === "range") {
                        const args = right.childForFieldName("arguments");
                        if (args && args.namedChildCount > 0) {
                            const argNodes = args.namedChildren;
                            const leftText = left ? left.text : "?";
                            if (argNodes.length === 1) {
                                condText = `${leftText} = 0, ${endof(argNodes[0])}, 1`;
                            } else if (argNodes.length === 2) {
                                condText = `${leftText} = ${argNodes[0].text}, ${endof(argNodes[1])}, 1`;
                            } else if (argNodes.length >= 3) {
                                condText = `${leftText} = ${argNodes[0].text}, ${endof(argNodes[1])}, ${argNodes[2].text}`;
                            }
                        }
                    }
                }
                
                return {
                    kind: "for",
                    cond: oneline(condText),
                    body: block(s.childForFieldName("body")),
                    else: block(s.childForFieldName("alternative"))
                };
            }
            if (isCpp) {
                const init = s.childForFieldName("initializer");
                const cond = s.childForFieldName("condition");
                const upd = s.childForFieldName("update");
                const body = s.childForFieldName("body");
                let spec = "";
                if (init) {
                    if (init.type === "declaration") {
                        let decls = [];
                        for (let i = 0; i < init.namedChildCount; i++) {
                            if (init.namedChildren[i].type === "init_declarator") decls.push(init.namedChildren[i].text);
                        }
                        spec += decls.length > 0 ? decls.join(", ") : init.text.replace(";", "");
                    } else {
                        spec += init.text.replace(";", "");
                    }
                }
                if (cond) spec += (spec ? ", " : "") + cond.text.replace(";", "");
                if (upd) spec += (spec ? ", " : "") + upd.text;
                return {
                    kind: "for",
                    cond: oneline(spec),
                    body: block(body),
                    else: block(null)
                };
            }
        }
        
        if (s.type === "do_statement") {
            const cond = s.childForFieldName("condition");
            const body = s.childForFieldName("body");
            let condText = cond ? cond.text : "?";
            if (isCpp && condText.startsWith("(") && condText.endsWith(")")) {
                condText = condText.substring(1, condText.length - 1);
            }
            return {
                kind: "dowhile",
                cond: oneline(condText),
                body: block(body)
            };
        }
        
        if (s.type === "switch_statement") {
            const cond = s.childForFieldName("condition");
            const body = s.childForFieldName("body");
            let subj = cond ? cond.text : "?";
            if (isCpp && subj.startsWith("(") && subj.endsWith(")")) {
                subj = subj.substring(1, subj.length - 1);
            }

            let cases = [];
            if (body && (body.type === "compound_statement" || body.type === "block")) {
                for (let i = 0; i < body.namedChildCount; i++) {
                    const c = body.namedChildren[i];
                    if (c.type === "case_statement") {
                         const valNode = c.childForFieldName("value");
                         cases.push({ val: valNode ? valNode.text : null, node: c });
                    }
                }
            }
            
            function buildCascade(i) {
                if (i >= cases.length) return { kind: "block", stmts: [] };
                const c = cases[i];
                const condText = c.val ? `${subj} == ${c.val}` : "";
                
                const caseStmts = [];
                for (let j = 0; j < c.node.namedChildCount; j++) {
                     const child = c.node.namedChildren[j];
                     if (child !== c.node.childForFieldName("value")) {
                          const mapped = stmt(child);
                          if (mapped) caseStmts.push(mapped);
                     }
                }
                const caseBlock = { kind: "block", stmts: caseStmts };
                
                if (condText === "") return caseBlock;
                return {
                     kind: "if",
                     cond: oneline(condText.replace(";", "")),
                     then: caseBlock,
                     else: buildCascade(i + 1)
                };
            }
            return buildCascade(0);
        }

        if (s.type === "match_statement") {
            const subjNode = s.childForFieldName("subject");
            const subj = subjNode ? subjNode.text : "?";
            const body = s.childForFieldName("body");
            
            let cases = [];
            if (body && body.type === "block") {
                for (let i = 0; i < body.namedChildCount; i++) {
                     const c = body.namedChildren[i];
                     if (c.type === "case_clause") {
                          const pat = c.namedChildren.find(n => n.type === "case_pattern");
                          const conseq = c.childForFieldName("consequence") || c.childForFieldName("body") || c.namedChildren.find(n => n.type === "block");
                          cases.push({ pat: pat ? pat.text : null, body: conseq });
                     }
                }
            }
            
            function buildCascade(i) {
                if (i >= cases.length) return { kind: "block", stmts: [] };
                const c = cases[i];
                let condText = c.pat ? `${subj} == ${c.pat}` : "";
                if (c.pat === "_" || c.pat === "case _") condText = "";
                
                if (condText === "") return block(c.body);
                return {
                     kind: "if",
                     cond: oneline(condText),
                     then: block(c.body),
                     else: buildCascade(i + 1)
                };
            }
            return buildCascade(0);
        }
        
        if (s.type === "try_statement") {
            const bodyNode = s.childForFieldName("body");
            const outStmts = block(bodyNode).stmts;
            
            let handlers = [];
            let finBlock = null;
            
            for (let i = 0; i < s.namedChildCount; i++) {
                const c = s.namedChildren[i];
                if (c.type === "except_clause" || c.type === "catch_clause") {
                     handlers.push(c);
                }
                if (c.type === "finally_clause") {
                     finBlock = c;
                }
            }
            
            function buildHandlers(i) {
                if (i >= handlers.length) return { kind: "block", stmts: [] };
                const h = handlers[i];
                let exType = "";
                if (h.type === "except_clause") {
                     const val = h.namedChildren.find(n => n.type !== "block");
                     if (val) exType = " " + val.text;
                } else if (h.type === "catch_clause") {
                     const p = h.childForFieldName("parameters");
                     if (p) {
                         let text = p.text;
                         if (text.startsWith("(") && text.endsWith(")")) text = text.substring(1, text.length - 1);
                         exType = " " + text;
                     }
                }
                
                let hBody = h.childForFieldName("body");
                if (!hBody) hBody = h.namedChildren[h.namedChildCount - 1];
                
                return {
                     kind: "block",
                     stmts: [{
                          kind: "if",
                          cond: oneline(`Виняток${exType}?`),
                          then: block(hBody),
                          else: buildHandlers(i + 1)
                     }]
                };
            }
            
            if (handlers.length > 0) {
                 outStmts.push(...buildHandlers(0).stmts);
            }
            
            // finally
            if (finBlock) {
                 const finBody = finBlock.namedChildren[finBlock.namedChildCount - 1];
                 outStmts.push(...block(finBody).stmts);
            }
            
            return { kind: "block", stmts: outStmts };
        }

        if (s.type === "with_statement") {
            const bodyNode = s.childForFieldName("body");
            const items = [];
            for (let i = 0; i < s.namedChildCount; i++) {
                 const c = s.namedChildren[i];
                 if (c.type === "with_item") items.push(c.text);
                 else if (c !== bodyNode && c.type !== "with_item" && c.type !== "block" && c.type !== "compound_statement") items.push(c.text);
            }
            const stmts = [{ kind: "process", text: oneline("відкрити: " + items.join(", ")) }];
            if (bodyNode) stmts.push(...block(bodyNode).stmts);
            return { kind: "block", stmts: stmts };
        }

        if (s.type === "break_statement") return { kind: "break" };
        if (s.type === "continue_statement") return { kind: "continue" };
        
        if (s.type === "return_statement") {
            let val = "";
            const vals = s.namedChildren.filter(c => c.type !== "comment");
            if (vals.length > 0) val = vals.map(c => c.text).join(" ");
            return {
                kind: "terminal",
                text: val ? oneline("Повернути " + val.replace(";", "")) : "Повернути"
            };
        }
        
        if (s.type === "raise_statement" || s.type === "throw_statement") {
            const cause = s.namedChildren.length > 0 ? s.namedChildren[0].text : "";
            return { kind: "terminal", text: oneline("Помилка" + (cause ? ": " + cause : "")) };
        }
        
        if (s.type === "expression_statement") {
            const expr = s.namedChildren[0];
            if (!expr) return { kind: "process", text: oneline(s.text.replace(";", "")) };
            // голий рядок-літерал (докстрінг/коментар) — не малюємо (як is_docstring у parser.py)
            if (expr.type === "string" || expr.type === "concatenated_string") return null;

            if (isCpp && (expr.type === "binary_expression" || expr.type === "shift_expression")) {
                let text = expr.text.trim();
                if (text.startsWith("cout") || text.startsWith("std::cout")) {
                    let ioText = text.replace(/^(?:std::)?cout\s*<<\s*/, "").replace(/<<\s*(?:std::)?endl$/, "").replace(/<<\s*(?:std::)?endl\s*/, "").replace(/<<\s*/g, " ");
                    ioText = ioText.replace(/"([^"]*)"/g, "«$1»");
                    return { kind: "io", text: oneline("Вивід " + ioText.replace(";", "")) };
                }
                if (text.startsWith("cin") || text.startsWith("std::cin")) {
                    let ioText = text.replace(/^(?:std::)?cin\s*>>\s*/, "").replace(/>>\s*/g, ", ");
                    return { kind: "io", text: oneline("Ввід " + ioText.replace(";", "")) };
                }
            }

            if (expr.type === "call" || expr.type === "call_expression") {
                const funcName = getCallName(expr);
                if (funcName === "print" || funcName === "cout" || funcName === "std::cout") {
                    const args = expr.childForFieldName("arguments");
                    if (args && args.namedChildCount > 0) {
                        const formattedArgs = args.namedChildren.map(formatArg).join(" ");
                        return { kind: "io", text: oneline("Вивід " + formattedArgs) };
                    }
                    return { kind: "io", text: "Вивід порожнього рядка" };
                }
                if (funcName === "input" || funcName === "cin" || funcName === "std::cin") {
                    const args = expr.childForFieldName("arguments");
                    if (args && args.namedChildCount > 0) {
                        const formattedArgs = args.namedChildren.map(formatArg).join(" ");
                        return { kind: "io", text: oneline("Ввід " + formattedArgs) };
                    }
                    return { kind: "io", text: "Ввід" };
                }
                if (funcName === "exit" || funcName === "quit") {
                    return { kind: "terminal", text: "Вихід" };
                }
                if (defined.has(funcName)) {
                    return { kind: "call", text: oneline(expr.text.replace(";", "")) };
                }
            }
            
            if (expr.type === "assignment" || expr.type === "assignment_expression") {
                if (hasInputNode(expr.childForFieldName("right"))) {
                    const left = expr.childForFieldName("left");
                    if (left) return { kind: "io", text: oneline("Ввід " + left.text.replace(";", "")) };
                }
                const rhs = expr.childForFieldName("right");
                if (rhs && (rhs.type === "call" || rhs.type === "call_expression")) {
                     const funcName = getCallName(rhs);
                     if (defined.has(funcName)) return { kind: "call", text: oneline(expr.text.replace(";", "")) };
                }
            }
            
            return { kind: "process", text: oneline(expr.text.replace(";", "")) };
        }
        
        if (s.type === "block" || s.type === "compound_statement") {
            return block(s);
        }

        if (s.type === "declaration") {
            if (hasInputNode(s)) {
                 const decl = s.childForFieldName("declarator");
                 if (decl && decl.childForFieldName("declarator")) {
                     return { kind: "io", text: oneline("Ввід " + decl.childForFieldName("declarator").text.replace(";", "")) };
                 }
                 return { kind: "io", text: oneline("Ввід " + s.text.replace(";", "")) };
            }
            
            if (isCpp) {
                let decls = [];
                for (let i = 0; i < s.namedChildCount; i++) {
                    if (s.namedChildren[i].type === "init_declarator") {
                        decls.push(s.namedChildren[i].text);
                    }
                }
                if (decls.length > 0) {
                    return { kind: "process", text: oneline(decls.join(", ").replace(/;/g, "")) };
                }
                return null;
            }

            return { kind: "process", text: oneline(s.text.replace(";", "")) };
        }

        if (s.isNamed) {
            if (s.type === "function_definition" || s.type === "class_definition" || 
                s.type === "import_statement" || s.type === "import_from_statement" || 
                s.type === "preproc_include" || s.type === "preproc_def" ||
                s.type === "comment" || s.type === "string") return null;

            return { kind: "process", text: oneline(s.text.replace(";", "")) };
        }
        return null;
    }

    function block(node) {
        if (!node) return { kind: "block", stmts: [] };
        let stmts = [];
        
        let children = node.type === "block" || node.type === "compound_statement" 
            ? node.namedChildren 
            : [node];

        for (let c of children) {
            const mapped = stmt(c);
            if (mapped) stmts.push(mapped);
        }
        return { kind: "block", stmts: stmts };
    }

    const funcs = [];
    const mainStmts = [];
    
    function collect(node) {
        if (!node) return;
        if (node.type === "function_definition") {
            let nameNode = node.childForFieldName("name");
            let paramsText = "";
            let bodyNode = node.childForFieldName("body");
            
            if (isCpp) {
                let decl = node.childForFieldName("declarator");
                while (decl && decl.type !== "identifier" && decl.type !== "function_declarator") {
                    decl = decl.childForFieldName("declarator");
                }
                if (decl && decl.type === "function_declarator") {
                    let n = decl.childForFieldName("declarator");
                    if (n && n.type === "identifier") nameNode = n;
                    let p = decl.childForFieldName("parameters");
                    if (p) {
                        let paramsList = [];
                        for (let i = 0; i < p.namedChildCount; i++) {
                            const paramDecl = p.namedChildren[i];
                            if (paramDecl.type === "parameter_declaration" || paramDecl.type === "optional_parameter_declaration") {
                                const d = paramDecl.childForFieldName("declarator");
                                if (d) paramsList.push(d.text);
                                else paramsList.push(paramDecl.text);
                            } else {
                                paramsList.push(paramDecl.text);
                            }
                        }
                        paramsText = paramsList.join(", ");
                    }
                }
            } else if (isPy) {
                let p = node.childForFieldName("parameters");
                if (p) {
                    // лише імена параметрів, без тип-анотацій і значень за замовчуванням
                    // (як [a.arg for a in fn.args.args] у parser.py): «matrix», не «matrix: T».
                    const names = [];
                    for (const pc of p.namedChildren) {
                        if (pc.type === "comment") continue;
                        if (pc.type === "identifier") { names.push(pc.text); continue; }
                        const nm = pc.childForFieldName("name");
                        if (nm) { names.push(nm.text); continue; }
                        const id = pc.namedChildren.find(c => c.type === "identifier");
                        names.push(id ? id.text : pc.text);
                    }
                    paramsText = names.join(", ");
                }
            }
            
            const fnName = nameNode ? nameNode.text : "unknown";
            const b = block(bodyNode);
            
            if (paramsText && paramsText.trim() !== "") {
                b.stmts.unshift({ kind: "io", text: "Ввід " + oneline(paramsText) });
            }
            
            funcs.push({ name: fnName, block: b });

            // вкладені функції — кожна окремою схемою (як collect(s.body) у parser.py).
            // block() уже прибрав їх із тіла батька; тут лише додаємо як власні фігури.
            if (bodyNode && bodyNode.type === "block") {
                for (const k of bodyNode.namedChildren) {
                    if (k.type === "function_definition") collect(k);
                }
            }
        } else {
            const mapped = stmt(node);
            if (mapped && mapped.kind !== "block") {
                 mainStmts.push(mapped);
            } else if (mapped && mapped.kind === "block") {
                 mainStmts.push(...mapped.stmts);
            }
        }
    }

    for (let c of tree.rootNode.namedChildren) {
        collect(c);
    }
    
    const out = [];
    for (let f of funcs) {
        out.push(f);
    }
    
    if (mainStmts.length > 0) {
        out.push({ name: defined.has("main") ? "програма" : "main", block: { kind: "block", stmts: mainStmts } });
    }
    
    if (out.length === 0) {
        out.push({ name: "main", block: { kind: "block", stmts: [] } });
    }
    
    return JSON.stringify(out);
}
